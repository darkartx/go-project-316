package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AnalizerAnalize_Page(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient, err := localClient(server.URL)
		if err != nil {
			t.Fatalf("localClient error: %v", err)
		}

		rootUrl, err := url.Parse(server.URL + "/about")
		if err != nil {
			t.Fatalf("parse rootUrl error: %v", err)
		}

		analizer := NewAnalizer(rootUrl, 10, httpClient)

		ctx := context.Background()

		startTime := time.Now()
		result, err := analizer.Analize(ctx)
		endTime := time.Now()

		if err != nil {
			t.Fatalf("analizer.Analize error: %v", err)
		}

		assert.Equal(t, server.URL+"/about", result.RootURL)
		assert.Equal(t, uint(10), result.Depth)
		assert.WithinRange(t, result.GeneratedAt, startTime, endTime)

		cases := map[string]struct {
			Depth          uint
			Error          string
			HTTPStatus     int
			BrokenLinksLen int
		}{
			server.URL + "/about": {
				Depth:      9,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/": {
				Depth:      8,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/contact": {
				Depth:      8,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/external-links": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/large-page": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/duplicate-links": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/mixed-content": {
				Depth:          7,
				HTTPStatus:     http.StatusOK,
				BrokenLinksLen: 1,
			},
			server.URL + "/nofollow-page": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/broken-links": {
				Depth:          7,
				HTTPStatus:     http.StatusOK,
				BrokenLinksLen: 5,
			},
			server.URL + "/redirect": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/anchor-links": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/deep/nested/page": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/no-links": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/form-page": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/blog": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/relative-links": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/robots.txt": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/sitemap.xml": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/blog?page=2&sort=date": {
				Depth:      6,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/blog/post-2": {
				Depth:      6,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/blog/post-1": {
				Depth:      6,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/not-found": {
				Depth:      6,
				HTTPStatus: http.StatusNotFound,
			},
			server.URL + "/this-does-not-exist": {
				Depth:      6,
				HTTPStatus: http.StatusNotFound,
			},
			server.URL + "/another-missing-page": {
				Depth:      6,
				HTTPStatus: http.StatusNotFound,
			},
			server.URL + "/api/data?format=json": {
				Depth:      6,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/server-error": {
				Depth:      6,
				HTTPStatus: http.StatusInternalServerError,
			},
		}

		for _, resultPage := range result.Pages {
			tt, ok := cases[resultPage.URL]
			if ok {
				delete(cases, resultPage.URL)
			} else {
				t.Errorf("unexpected result page: %s", resultPage.URL)
				continue
			}

			assert.Equal(t, tt.Depth, resultPage.Depth)
			assert.Equal(t, tt.Error, resultPage.Error)
			assert.Equal(t, tt.HTTPStatus, resultPage.HTTPStatus)
			assert.Len(t, resultPage.BrokenLinks, tt.BrokenLinksLen)
			assert.WithinRange(t, resultPage.DiscoveredAt, startTime, endTime)
		}

		for url, _ := range cases {
			t.Errorf("expected result page: %s", url)
		}
	})
}

func Test_AnalizerAnalize_CtxWithTimeout(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient, err := localClient(server.URL)
		if err != nil {
			t.Fatalf("localClient error: %v", err)
		}

		rootUrl, err := url.Parse(server.URL + "/")
		if err != nil {
			t.Fatalf("parse rootUrl error: %v", err)
		}

		analizer := NewAnalizer(rootUrl, 2, httpClient)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)

		startTime := time.Now()
		result, err := analizer.Analize(ctx)
		endTime := time.Now()

		cancel()

		if err != nil {
			t.Fatalf("analizer.Analize error: %v", err)
		}

		assert.Equal(t, server.URL+"/", result.RootURL)
		assert.Equal(t, uint(2), result.Depth)
		assert.WithinRange(t, result.GeneratedAt, startTime, endTime)
	})
}
