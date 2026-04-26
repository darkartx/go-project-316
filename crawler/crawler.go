package crawler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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
	httpClient := opts.HTTPClient
	httpClient.Timeout = opts.Timeout

	rootUrl, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}

	analizer := NewAnalizer(rootUrl, opts.Depth, httpClient)
	report, err := analizer.Analize(ctx)
	if err != nil {
		return nil, err
	}

	return makeJsonReport(report)
}

func makeJsonReport(report Report) ([]byte, error) {
	result, err := json.MarshalIndent(report, "", "  ")

	if err != nil {
		return nil, err
	}

	return result, nil
}
