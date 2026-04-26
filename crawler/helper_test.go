package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

const fixturesDir = "testdata/fixtures"

func withTestServer(t *testing.T, fn func(server *httptest.Server)) {
	t.Helper()
	server := setupTestServer(t)

	fn(server)
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Main index page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveFixture(t, w, "index.html", "text/html")
	})

	// About page
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "about.html", "text/html")
	})

	// Contact page
	mux.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "contact.html", "text/html")
	})

	// Page with deep nested links
	mux.HandleFunc("/deep/nested/page", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "deep_nested.html", "text/html")
	})

	// Page with external links only
	mux.HandleFunc("/external-links", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "external_links.html", "text/html")
	})

	// Page with broken links
	mux.HandleFunc("/broken-links", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "broken_links.html", "text/html")
	})

	// Page that returns 404
	mux.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Page that returns 500
	mux.HandleFunc("/server-error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	// Page that redirects
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/about", http.StatusMovedPermanently)
	})

	// Page with redirect loop
	mux.HandleFunc("/redirect-loop-a", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect-loop-b", http.StatusFound)
	})
	mux.HandleFunc("/redirect-loop-b", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect-loop-a", http.StatusFound)
	})

	// Page with no links
	mux.HandleFunc("/no-links", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "no_links.html", "text/html")
	})

	// Page with duplicate links
	mux.HandleFunc("/duplicate-links", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "duplicate_links.html", "text/html")
	})

	// Page with anchor links
	mux.HandleFunc("/anchor-links", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "anchor_links.html", "text/html")
	})

	// Page with form
	mux.HandleFunc("/form-page", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "form_page.html", "text/html")
	})

	// CSS assets
	mux.HandleFunc("/assets/css/main.css", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/css/main.css", "text/css")
	})
	mux.HandleFunc("/assets/css/responsive.css", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/css/responsive.css", "text/css")
	})

	// JavaScript assets
	mux.HandleFunc("/assets/js/app.js", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/js/app.js", "application/javascript")
	})
	mux.HandleFunc("/assets/js/vendor.js", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/js/vendor.js", "application/javascript")
	})

	// Image assets
	mux.HandleFunc("/assets/images/logo.png", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/images/logo.png", "image/png")
	})
	mux.HandleFunc("/assets/images/banner.jpg", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/images/banner.jpg", "image/jpeg")
	})
	mux.HandleFunc("/assets/images/icon.svg", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/images/icon.svg", "image/svg+xml")
	})

	// Font assets
	mux.HandleFunc("/assets/fonts/font.woff2", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "assets/fonts/font.woff2", "font/woff2")
	})

	// JSON API endpoint
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "api/data.json", "application/json")
	})

	// XML sitemap
	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "sitemap.xml", "application/xml")
	})

	// Robots.txt
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "robots.txt", "text/plain")
	})

	// Page with nofollow links
	mux.HandleFunc("/nofollow-page", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "nofollow_page.html", "text/html")
	})

	// Page with relative links
	mux.HandleFunc("/relative-links", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "relative_links.html", "text/html")
	})

	// Page with mixed content
	mux.HandleFunc("/mixed-content", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "mixed_content.html", "text/html")
	})

	// Paginated pages
	mux.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "blog/index.html", "text/html")
	})
	mux.HandleFunc("/blog/page/2", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "blog/page2.html", "text/html")
	})
	mux.HandleFunc("/blog/post-1", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "blog/post1.html", "text/html")
	})
	mux.HandleFunc("/blog/post-2", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "blog/post2.html", "text/html")
	})

	// Slow response page (for timeout testing)
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "slow.html", "text/html")
	})

	// Large page
	mux.HandleFunc("/large-page", func(w http.ResponseWriter, r *http.Request) {
		serveFixture(t, w, "large_page.html", "text/html")
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

func serveFixture(t *testing.T, w http.ResponseWriter, filename string, contentType string) {
	t.Helper()
	path := filepath.Join(fixturesDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Logf("fixture not found: %s, serving empty response", path)
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func localClient(serverURL string) (*http.Client, error) {
	serverURLParsed, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &localClientTransport{serverURLParsed},
	}

	return client, nil
}

type localClientTransport struct {
	localURL *url.URL
}

func (t *localClientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == t.localURL.Host && req.URL.Port() == t.localURL.Port() {
		return http.DefaultTransport.RoundTrip(req)
	}

	return &http.Response{
		Status:        http.StatusText(http.StatusOK),
		StatusCode:    http.StatusOK,
		Body:          http.NoBody,
		ContentLength: 0,
	}, nil
}
