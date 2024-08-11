package uboa

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"gonum.org/v1/gonum/stat"
)

func (b *Uboa) Send(ctx context.Context, outChan chan<- *Metrics) {
	req, err := http.NewRequest(b.Method, b.URL, bytes.NewBufferString(b.Body))
	if err != nil {
		outChan <- &Metrics{
			Error: fmt.Sprintf("Error creating requests: %v", err),
		}
		return
	}

	for k, v := range b.Headers {
		req.Header.Add(k, v)
	}

	var t0, t1, t2, t3, t4 time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t0.IsZero() {
				// connecting directly to IP
				t1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				outChan <- &Metrics{
					Error: fmt.Sprintf("unable to connect to host %v: %v", addr, err),
				}
				return
			}
			t2 = time.Now()
		},
		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
		// Todo: Add tls support
		// TLSHandshakeStart:    func() { t5 = time.Now() },
		// TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6 = time.Now() },
	}

	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))
	var resp *http.Response
	maxRetries := b.MaxRetries
	baseDelay := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		resp, err = b.Client.Do(req)
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			time.Sleep(baseDelay * time.Duration(1<<uint(i))) // Exponential backoff
		}
	}

	if err != nil {
		var errr string
		var status int
		ue, ok := err.(*url.Error)
		switch {
		case resp != nil:
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				errr = fmt.Sprint("Failed to read HTTP response body", err)
			}
			resp.Body.Close()
			status = resp.StatusCode
		case ok && ue.Err == context.DeadlineExceeded:
			errr = "Timeout"
		case ok && ue.Err == context.Canceled:
			errr = "Cancelled"
		case ok:
			errr = ue.Err.Error()
		default:
			errr = err.Error()
		}

		outChan <- &Metrics{
			Error:      errr,
			StatusCode: status,
		}
		return
	}

	respSize, _ := io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	b.Bar.Add(1)
	t5 := time.Now()

	out := &Metrics{
		DNSLookup:        float64(t1.Sub(t0).Milliseconds()),
		TCPConn:          float64(t2.Sub(t1).Milliseconds()),
		ServerProcessing: float64(t4.Sub(t3).Milliseconds()),
		ContentTransfer:  float64(t5.Sub(t4).Milliseconds()),
		StatusCode:       resp.StatusCode,
		RespSize:         respSize,
	}

	out.RespDuration = out.DNSLookup + out.TCPConn + out.ServerProcessing + out.ContentTransfer

	outChan <- out

}

func (b *Uboa) Load() *ResultMetrics {
	var tcpDur []float64
	var respDur []float64
	var serverDur []float64
	var transferDur []float64
	var failedRequests int
	var totalRespSize int64

	ag := make(map[string]AggregateMetrics)
	s := &SummaryMetrics{
		TotalRequests: b.Requests,
		StatusCodes:   make(map[int]int),
	}

	b.Bar = progressbar.NewOptions(b.Requests,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[_cyan_] [reset]",
			SaucerHead:    "[_cyan_] [reset]",
			SaucerPadding: " ",
			BarStart:      "[light_cyan][",
			BarEnd:        "[light_cyan]]",
		}))
	clor := color.New(color.FgHiCyan).Add(color.Italic)
	clor.Printf("\nStarting Load Test with %d requests using %d concurrent users\n\n", b.Requests, b.Clients)

	b.Client = &http.Client{
		Timeout: time.Second * time.Duration(b.Timeout),
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10000,
			DisableCompression:  false,
			DisableKeepAlives:   b.DisableKeepAlives, // Enable keep-alives
		},
	}

	var wg sync.WaitGroup
	m := make(chan *Metrics, b.Requests)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	ctx := context.Background()

	reqPerClient := b.Requests / b.Clients
	extraReq := b.Requests % b.Clients

	wg.Add(b.Clients)
	start := time.Now()

	for i := 0; i < b.Clients; i++ {
		go func(id int) {
			defer wg.Done()
			numReq := reqPerClient
			if id < extraReq {
				numReq++
			}
			for j := 0; j < numReq; j++ {
				b.Send(ctx, m)
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(m)
	}()

	prev := time.Now()
	for mm := range m {
		if mm.TCPConn < 1000 {
			tcpDur = append(tcpDur, mm.TCPConn)
		}
		serverDur = append(serverDur, mm.ServerProcessing)
		transferDur = append(transferDur, mm.ContentTransfer)
		respDur = append(respDur, mm.RespDuration)
		if mm.StatusCode > 226 {
			failedRequests++
		}

		if mm.StatusCode >= 100 {
			s.StatusCodes[mm.StatusCode]++
		}

		totalRespSize += mm.RespSize

		if currTime := time.Now(); currTime.Sub(prev) > time.Second {
			t := time.Now().Format("15:04:05")
			var a AggregateMetrics
			if len(tcpDur) != 0 {
				// TCPConn
				slices.Sort(tcpDur)
				a.TCPConnStat.Mean = stat.Mean(tcpDur, nil)
				a.TCPConnStat.P90 = stat.Quantile(0.9, stat.LinInterp, tcpDur, nil)
				a.TCPConnStat.P95 = stat.Quantile(0.95, stat.LinInterp, tcpDur, nil)
				a.TCPConnStat.P99 = stat.Quantile(0.99, stat.LinInterp, tcpDur, nil)
			}

			if len(transferDur) != 0 {
				// ContentTransfer
				slices.Sort(transferDur)
				a.ContentTransferStat.Mean = stat.Mean(transferDur, nil)
				a.ContentTransferStat.P90 = stat.Quantile(0.9, stat.LinInterp, transferDur, nil)
				a.ContentTransferStat.P95 = stat.Quantile(0.95, stat.LinInterp, transferDur, nil)
				a.ContentTransferStat.P99 = stat.Quantile(0.99, stat.LinInterp, transferDur, nil)
			}

			if len(serverDur) != 0 {
				// ServerProcessing
				slices.Sort(serverDur)
				a.ServerProcessingStat.Mean = stat.Mean(serverDur, nil)
				a.ServerProcessingStat.P90 = stat.Quantile(0.9, stat.LinInterp, serverDur, nil)
				a.ServerProcessingStat.P95 = stat.Quantile(0.95, stat.LinInterp, serverDur, nil)
				a.ServerProcessingStat.P99 = stat.Quantile(0.99, stat.LinInterp, serverDur, nil)
			}

			if len(respDur) != 0 {
				// RespDuration
				slices.Sort(respDur)
				a.RespDurationStat.Mean = stat.Mean(respDur, nil)
				a.RespDurationStat.P90 = stat.Quantile(0.9, stat.LinInterp, respDur, nil)
				a.RespDurationStat.P95 = stat.Quantile(0.95, stat.LinInterp, respDur, nil)
				a.RespDurationStat.P99 = stat.Quantile(0.99, stat.LinInterp, respDur, nil)
			}

			ag[t] = a
			prev = time.Now()

		}
	}

	duration := time.Since(start).Seconds()
	s.RequestsPerSecond = float64(s.TotalRequests) / duration
	s.FailedRequests = failedRequests
	s.SuccessFulRequests = s.TotalRequests - failedRequests
	slices.Sort(respDur)
	s.AvgRespTime = stat.Mean(respDur, nil)
	s.RespSizePerSec = float64(totalRespSize) / duration

	return &ResultMetrics{
		*s,
		ag,
	}
}
