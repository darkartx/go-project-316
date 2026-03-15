package crawler

import (
	"context"
	"encoding/json"
	"net/http"
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

type Options struct {
	URL         string
	Depth       uint
	Retries     uint
	Delay       time.Duration
	Timeout     time.Duration
	UserAgent   string
	Concurrency uint
	IndentJSON  uint
	HTTPClient  *http.Client
}

func Analize(ctx context.Context, opts Options) ([]byte, error) {
	report, err := analize(ctx, opts)

	if err != nil {
		return nil, err
	}

	result, err := json.MarshalIndent(report, "", "  ")

	if err != nil {
		return nil, err
	}

	return result, nil
}

func analize(ctx context.Context, opts Options) (Report, error) {
	var report Report

	report.RootURL = opts.URL
	report.Depth = opts.Depth
	report.GeneratedAt = time.Now()
	report.Pages = make([]ReportPage, 1)

	report.Pages[0] = analizePage(report.RootURL, report.Depth, opts.HTTPClient)

	return report, nil
}

func analizePage(url string, depth uint, httpClient *http.Client) ReportPage {
	report := ReportPage{url, depth - 1, 0, "ok", ""}

	response, err := httpClient.Get(url)
	if err != nil {
		report.Error = err.Error()
		return report
	}

	report.HTTPStatus = response.StatusCode

	return report
}
