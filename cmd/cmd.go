package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	bam "github.com/oddegen/bam/internal/pkg"
	"github.com/oddegen/bam/ui"
	"github.com/urfave/cli/v2"
)

func Run() {
	app := &cli.App{
		Name:  "uboa",
		Usage: "A local first HTTP load testing CLI tool",
		Description: "uboa is a HTTP load testing tool designed to help you evaluate " +
			"the performance and reliability of\nyour web applications under various levels of concurrent traffic.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Usage:    "Target URL to test (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "method",
				Aliases: []string{"m"},
				Usage:   "HTTP method for requests (GET, POST, PUT, etc.)",
				Value:   "GET",
			},
			&cli.StringFlag{
				Name:    "headers",
				Aliases: []string{"H"},
				Usage:   "Custom HTTP headers (format: key1:value1,key2:value2)",
			},
			&cli.StringFlag{
				Name:    "body",
				Aliases: []string{"d"},
				Usage:   "Request body for POST, PUT, or PATCH requests",
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Output results in JSON format",
			},
			&cli.StringFlag{
				Name:        "html",
				Aliases:     []string{"html-output"},
				Usage:       "Generate an HTML report and save to specified file",
				Value:       time.Now().Format("2006-01-02") + "_upoa-result.html",
				DefaultText: "yyyy-mm-dd_upoa-result.html",
			},
			&cli.BoolFlag{
				Name:    "skip-preview",
				Aliases: []string{"S"},
				Usage:   "Skip automatic preview of results",
			},
			&cli.BoolFlag{
				Name:    "plain",
				Aliases: []string{"p"},
				Usage:   "Output results in plain, machine-readable format",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "File path for saving the output",
				Value:   "output",
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "Number of concurrent clients",
				Value:   5,
			},
			&cli.IntFlag{
				Name:    "requests",
				Aliases: []string{"n"},
				Usage:   "Total number of requests to send",
				Value:   100,
			},
			&cli.IntFlag{
				Name:    "timeout",
				Aliases: []string{"T"},
				Usage:   "HTTP client timeout in seconds",
				Value:   5,
			},
			&cli.BoolFlag{
				Name:    "keep-alive",
				Aliases: []string{"k"},
				Usage:   "Enable HTTP keep-alive connections",
			},
		},
		Action: validate,
	}
	setCustomCLITemplate(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func validate(c *cli.Context) error {
	var headerMap map[string]string
	urlstr := c.String("url")
	method := c.String("method")
	headers := c.String("headers")
	json := c.Bool("json")
	html := c.String("html")
	skip := c.Bool("skip-preview")
	plain := c.Bool("plain")
	outputFile := c.String("output")
	concurrency := c.Int("concurrency")
	requests := c.Int("requests")
	keepAlive := c.Bool("keep-alive")
	timeout := c.Int("timeout")

	if concurrency == 0 {
		return errors.New("error: Concurrency level cannot be set to: 0")
	}

	if requests == 0 {
		return errors.New("error: No. of request cannot be set to: 0")
	}

	if u, err := url.Parse(urlstr); !(err == nil && u.Scheme != "" && u.Host != "") {
		return errors.New("error: Not a valid URL. Must have the following format: http{s}://{host}")
	}

	if headers != "" {
		header := strings.Split(headers, ",")
		headerMap := make(map[string]string)
		for _, h := range header {
			k, v, ok := strings.Cut(h, ":")
			if !ok {
				return errors.New("error: Not a valid header value")
			}
			headerMap[k] = v
		}
	}

	if timeout < 0 {
		return errors.New("error: Timeout cannot be less than 0")
	}

	if !isValidHttpMethod(method) {
		return errors.New("error: invalid HTTP method")
	}

	b := &bam.Bam{
		URL:               urlstr,
		Method:            method,
		Headers:           headerMap,
		ExportJson:        json,
		Plain:             plain,
		HtmlFile:          html,
		Clients:           concurrency,
		Requests:          requests,
		OutputFile:        outputFile,
		DisableKeepAlives: keepAlive,
		Timeout:           time.Duration(timeout),
	}

	a := b.Load()
	b.Template = &bam.Template{
		Result: a,
	}

	table := uitable.New()
	table.MaxColWidth = 80

	table.AddRow("Total Requests:", fmt.Sprintf("%d reqs", a.TotalRequests))
	table.AddRow("Average Response Time:", fmt.Sprintf("%.2f ms", a.AvgRespTime))
	table.AddRow("Error rate:", fmt.Sprintf("%.2f%%", a.ErrorPercentage))
	table.AddRow("Request Per Second:", fmt.Sprintf("%.2f req/s", a.RequestsPerSecond))
	fmt.Println("\n\nSummary:")
	fmt.Println(table)

	outPutJSON(b.OutputFile, a)
	err := ui.Render(*a, b.HtmlFile, skip)
	return err
}

func outPutJSON(fileName string, metrics *bam.ResultMetrics) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc, _ := json.MarshalIndent(&metrics, "", "  ")
	_, err = f.Write(enc)
	return err
}

func isValidHttpMethod(method string) bool {
	httpMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
	for _, m := range httpMethods {
		if m == strings.ToUpper(method) {
			return true
		}
	}
	return false
}

func setCustomCLITemplate(c *cli.App) {
	whiteBold := color.New(color.Bold).SprintfFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	c.CustomAppHelpTemplate = fmt.Sprintf(
		`%s:
	{{.Name}}{{if .Usage}} - {{.Usage}}{{end}}{{if .Description}}

%s:
{{indent 2 .Description }}{{end}}{{if .VisibleCommands}}

%s:{{range .VisibleCategories}}{{if .Name}}
	{{.Name}}:{{range .VisibleCommands}}
	  {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
	{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

%s:
	{{range $index, $option := .VisibleFlags}}{{if $index}}
	{{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
	{{.Copyright}}{{end}}

	Example of running uboa with 100 requests using 10 concurrent users
	%s
  `, whiteBold("NAME"),
		whiteBold("DESCRIPTION"),
		whiteBold("COMMANDS"),
		whiteBold("GLOBAL OPTIONS"),
		cyan("$ uboa run -u https://dummyjson.com/auth/RESOURCE -c 10 -n 100"))
}
