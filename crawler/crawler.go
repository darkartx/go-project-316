package crawler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
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

	analizer := NewAnalizer(rootUrl, opts.Depth, opts.Retries, opts.UserAgent, httpClient)
	report, err := analizer.Analize(ctx)
	if err != nil {
		return nil, err
	}

	return makeJsonReport(report, opts.IndentJSON)
}

func makeJsonReport(report Report, indentSize uint) ([]byte, error) {
	indent := strings.Repeat(" ", int(indentSize))
	result, err := json.MarshalIndent(report, "", indent)

	if err != nil {
		return nil, err
	}

	return result, nil
}
