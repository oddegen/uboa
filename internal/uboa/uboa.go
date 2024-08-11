package uboa

import (
	"net/http"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Uboa struct {
	URL                    string
	Method                 string
	Body                   string
	Headers                map[string]string
	OutputFileName         string
	ExportJson, ExportHtml bool
	Clients                int
	Requests               int
	Timeout                time.Duration
	Bar                    *progressbar.ProgressBar
	Template               *Template
	Client                 *http.Client
	DisableKeepAlives      bool
	MaxRetries             int
}
