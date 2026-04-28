package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AnalizerAnalize_BrokenLinks(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient, err := localClient(server.URL)
		if err != nil {
			t.Fatalf("localClient error: %v", err)
		}

		rootUrl, err := url.Parse(server.URL + "/broken-links")
		if err != nil {
			t.Fatalf("parse rootUrl error: %v", err)
		}

		analizer := NewAnalizer(rootUrl, 1, httpClient)

		ctx := context.Background()

		startTime := time.Now()
		result, err := analizer.Analize(ctx)
		endTime := time.Now()

		if err != nil {
			t.Fatalf("analizer.Analize error: %v", err)
		}

		assert.Equal(t, server.URL+"/broken-links", result.RootURL)
		assert.Equal(t, uint(1), result.Depth)
		assert.WithinRange(t, result.GeneratedAt, startTime, endTime)
		assert.Equal(t, uint(0), result.Pages[0].Depth)
		assert.Equal(t, "", result.Pages[0].Error)
		assert.Equal(t, http.StatusOK, result.Pages[0].HTTPStatus)
		assert.Equal(t, "ok", result.Pages[0].Status)
		assert.Equal(t, server.URL+"/broken-links", result.Pages[0].URL)

		cases := map[string]struct {
			Error      string
			StatusCode int
		}{
			server.URL + "/not-found": {
				Error:      "",
				StatusCode: http.StatusNotFound,
			},
			server.URL + "/server-error": {
				Error:      "",
				StatusCode: http.StatusInternalServerError,
			},
			server.URL + "/this-does-not-exist": {
				Error:      "",
				StatusCode: http.StatusNotFound,
			},
			server.URL + "/another-missing-page": {
				Error:      "",
				StatusCode: http.StatusNotFound,
			},
			server.URL + "/assets/images/missing.png": {
				Error:      "",
				StatusCode: http.StatusNotFound,
			},
		}

		for _, brokenLink := range result.Pages[0].BrokenLinks {
			tt, ok := cases[brokenLink.URL]
			if ok {
				delete(cases, brokenLink.URL)
			} else {
				t.Errorf("unexpected broken link: %s", brokenLink.URL)
				continue
			}

			assert.Equal(t, tt.Error, brokenLink.Error)
			assert.Equal(t, tt.StatusCode, brokenLink.StatusCode)
		}

		for url := range cases {
			t.Errorf("expected broken link: %s", url)
		}
	})
}

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
			Depth             uint
			Error             string
			HTTPStatus        int
			BrokenLinksLen    int
			SeoHasTitle       bool
			SeoTitle          string
			SeoHasDescription bool
			SeoDescription    string
			SeoHasH1          bool
		}{
			server.URL + "/about": {
				Depth:             9,
				HTTPStatus:        http.StatusOK,
				SeoHasTitle:       true,
				SeoTitle:          "About Us",
				SeoHasDescription: true,
				SeoDescription:    "Test Site - Description",
				SeoHasH1:          true,
			},
			server.URL + "/": {
				Depth:             8,
				HTTPStatus:        http.StatusOK,
				SeoHasTitle:       true,
				SeoTitle:          "Test Site - Home&",
				SeoHasDescription: true,
				SeoDescription:    "Test Site - Description&",
				SeoHasH1:          true,
			},
			server.URL + "/contact": {
				Depth:       8,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Contact",
				SeoHasH1:    true,
			},
			server.URL + "/external-links": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "External Links Page",
				SeoHasH1:    true,
			},
			server.URL + "/large-page": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Large Page",
				SeoHasH1:    true,
			},
			server.URL + "/duplicate-links": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Duplicate Links Page",
				SeoHasH1:    true,
			},
			server.URL + "/mixed-content": {
				Depth:          7,
				HTTPStatus:     http.StatusOK,
				SeoHasTitle:    true,
				SeoTitle:       "Mixed Content Page",
				SeoHasH1:       true,
				BrokenLinksLen: 1,
			},
			server.URL + "/nofollow-page": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "NoFollow Page",
				SeoHasH1:    true,
			},
			server.URL + "/broken-links": {
				Depth:          7,
				HTTPStatus:     http.StatusOK,
				SeoHasTitle:    true,
				SeoTitle:       "Broken Links Page",
				SeoHasH1:       true,
				BrokenLinksLen: 5,
			},
			server.URL + "/redirect": {
				Depth:             7,
				HTTPStatus:        http.StatusOK,
				SeoHasTitle:       true,
				SeoTitle:          "About Us",
				SeoHasDescription: true,
				SeoDescription:    "Test Site - Description",
				SeoHasH1:          true,
			},
			server.URL + "/anchor-links": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Anchor Links Page",
				SeoHasH1:    true,
			},
			server.URL + "/deep/nested/page": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Deep Nested Page",
				SeoHasH1:    true,
			},
			server.URL + "/no-links": {
				Depth:      7,
				HTTPStatus: http.StatusOK,
			},
			server.URL + "/form-page": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Form Page",
				SeoHasH1:    true,
			},
			server.URL + "/blog": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Blog",
				SeoHasH1:    true,
			},
			server.URL + "/relative-links": {
				Depth:       7,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Relative Links Page",
				SeoHasH1:    true,
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
				Depth:       6,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Blog",
				SeoHasH1:    true,
			},
			server.URL + "/blog/post-2": {
				Depth:       6,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Blog Post 2",
				SeoHasH1:    true,
			},
			server.URL + "/blog/post-1": {
				Depth:       6,
				HTTPStatus:  http.StatusOK,
				SeoHasTitle: true,
				SeoTitle:    "Blog Post 1",
				SeoHasH1:    true,
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

			// fmt.Printf("~~~ %v\n", resultPage.URL)

			assert.Equal(t, tt.Depth, resultPage.Depth)
			assert.Equal(t, tt.Error, resultPage.Error)
			assert.Equal(t, tt.HTTPStatus, resultPage.HTTPStatus)
			assert.Equal(t, tt.SeoHasTitle, resultPage.Seo.HasTitle)
			assert.Equal(t, tt.SeoTitle, resultPage.Seo.Title)
			assert.Equal(t, tt.SeoHasDescription, resultPage.Seo.HasDescription)
			assert.Equal(t, tt.SeoDescription, resultPage.Seo.Description)
			assert.Equal(t, tt.SeoHasH1, resultPage.Seo.HasH1)
			assert.Len(t, resultPage.BrokenLinks, tt.BrokenLinksLen)
			assert.WithinRange(t, resultPage.DiscoveredAt, startTime, endTime)
		}

		for url := range cases {
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

		analizer := NewAnalizer(rootUrl, 10, httpClient)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)

		startTime := time.Now()
		result, err := analizer.Analize(ctx)
		endTime := time.Now()

		cancel()

		if err != nil {
			t.Fatalf("analizer.Analize error: %v", err)
		}

		assert.Equal(t, server.URL+"/", result.RootURL)
		assert.Equal(t, uint(10), result.Depth)
		assert.WithinRange(t, result.GeneratedAt, startTime, endTime)
	})
}

func Test_AnalizerAnalize_Timeout(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()

	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Microsecond

	rootUrl, err := url.Parse(server.URL)
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

	assert.Equal(t, server.URL, result.RootURL)
	assert.Equal(t, uint(10), result.Depth)
	assert.WithinRange(t, result.GeneratedAt, startTime, endTime)

	cases := map[string]struct {
		Depth      uint
		Error      string
		HTTPStatus int
	}{
		server.URL: {
			Depth:      9,
			HTTPStatus: 0,
			Error:      fmt.Sprintf(`Get "%s": context deadline exceeded (Client.Timeout exceeded while awaiting headers)`, server.URL),
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
		assert.WithinRange(t, resultPage.DiscoveredAt, startTime, endTime)
	}

	for url := range cases {
		t.Errorf("expected result page: %s", url)
	}
}

func Test_AnalizerAnalize_NetworkError(t *testing.T) {
	httpClient := failingClient()

	rootUrl, err := url.Parse("https://example.com")
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

	assert.Equal(t, "https://example.com", result.RootURL)
	assert.Equal(t, uint(10), result.Depth)
	assert.WithinRange(t, result.GeneratedAt, startTime, endTime)

	cases := map[string]struct {
		Depth      uint
		Error      string
		HTTPStatus int
	}{
		"https://example.com": {
			Depth:      9,
			HTTPStatus: 0,
			Error:      `Get "https://example.com": connection refused: network unreachable`,
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
		assert.WithinRange(t, resultPage.DiscoveredAt, startTime, endTime)
	}

	for url := range cases {
		t.Errorf("expected result page: %s", url)
	}
}
