package uboa

import (
	"encoding/json"
)

type ResultMetrics struct {
	SummaryMetrics      `json:"summary_metrics"`
	AggregateMetricsMap map[string]AggregateMetrics `json:"aggregate_metrics"`
}

type SummaryMetrics struct {
	TotalRequests     int         `json:"total_requests"`
	ErrorPercentage   float64     `json:"error_percentage"`
	RequestsPerSecond float64     `json:"requests_per_second"`
	StatusCodes       map[int]int `json:"status_codes"`
	AvgRespTime       float64     `json:"avg_resp_time"`
	RespSizePerSec    float64     `json:"resp_per_sec"`
}

type AggregateMetrics struct {
	TCPConnStat          Stat `json:"tcp_conn_stat"`
	ServerProcessingStat Stat `json:"server_processing_stat"`
	ContentTransferStat  Stat `json:"content_transfer_stat"`
	RespDurationStat     Stat `json:"resp_duration_stat"`
}

type Metrics struct {
	DNSLookup        float64 `json:"dns_lookup"`
	TCPConn          float64 `json:"tcp_conn"`
	ServerProcessing float64 `json:"server_processing"`
	ContentTransfer  float64 `json:"content_transfer"`
	RespDuration     float64 `json:"resp_duration"`
	StatusCode       int     `json:"status_code"`
	RespSize         int64   `json:"resp_size"`
	Error            string  `json:"error"`
}

type Stat struct {
	Mean float64 `json:"mean"`
	P90  float64 `json:"p90"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
}

func (rm *ResultMetrics) MarshalJSON() ([]byte, error) {
	embed := struct {
		Result struct {
			SummaryMetrics      `json:"summary_metrics"`
			AggregateMetricsMap map[string]AggregateMetrics `json:"aggregate_metrics"`
		} `json:"result"`
	}{}
	embed.Result.SummaryMetrics = rm.SummaryMetrics
	// aggregateMetricsMap := make(map[string]AggregateMetrics)
	// for k, v := range rm.AggregateMetricsMap {
	// 	aggregateMetricsMap[strconv.FormatFloat(k, 'f', 6, 64)] = v
	// }
	embed.Result.AggregateMetricsMap = rm.AggregateMetricsMap
	return json.Marshal(embed)
}
