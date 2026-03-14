package crawler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

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
	report.Pages[0].URL = opts.URL
	report.Pages[0].Depth = report.Depth - 1
	report.Pages[0].HTTPStatus = http.StatusOK
	report.Pages[0].Status = "ok"
	report.Pages[0].Error = ""

	return report, nil
}
