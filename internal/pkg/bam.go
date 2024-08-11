package bam

import (
	"net/http"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Bam struct {
	URL               string
	Method            string
	Body              string
	Headers           map[string]string
	OutputFile        string
	ExportJson, Plain bool
	HtmlFile          string
	Clients           int
	Requests          int
	Timeout           time.Duration
	Bar               *progressbar.ProgressBar
	Template          *Template
	Client            *http.Client
	DisableKeepAlives bool
}
