package crawler

import (
	"time"
)

type Report struct {
	RootURL     string       `json:"root_url"`
	Depth       uint         `json:"depth"`
	GeneratedAt time.Time    `json:"generated_at"`
	Pages       []ReportPage `json:"pages"`
}

type ReportPage struct {
	URL        string `json:"url"`
	Depth      uint   `json:"depth"`
	HTTPStatus int    `json:"http_status"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}
