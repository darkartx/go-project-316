package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AnalizeUrlGet_Page(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypePage, result.UrlType)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
		assert.True(t, result.PageData.Seo.HasTitle)
		assert.Equal(t, "Test Site - Home&", result.PageData.Seo.Title)
		assert.True(t, result.PageData.Seo.HasDescription)
		assert.Equal(t, "Test Site - Description&", result.PageData.Seo.Description)
		assert.True(t, result.PageData.Seo.HasH1)

		expectedLinks := []struct {
			Url      string
			LinkType LinkType
		}{
			{"/assets/css/main.css", LinkTypeAsset},
			{"/assets/css/responsive.css", LinkTypeAsset},
			{"/assets/images/icon.svg", LinkTypeAsset},
			{"/assets/images/logo.png", LinkTypeAsset},
			{"/", LinkTypePage},
			{"/about", LinkTypePage},
			{"/contact", LinkTypePage},
			{"/blog", LinkTypePage},
			{"/nofollow-page", LinkTypePage},
			{"/assets/images/banner.jpg", LinkTypeAsset},
			{"/deep/nested/page", LinkTypePage},
			{"/external-links", LinkTypePage},
			{"/broken-links", LinkTypePage},
			{"/no-links", LinkTypePage},
			{"/duplicate-links", LinkTypePage},
			{"/anchor-links", LinkTypePage},
			{"/form-page", LinkTypePage},
			{"/relative-links", LinkTypePage},
			{"/mixed-content", LinkTypePage},
			{"/redirect", LinkTypePage},
			{"/large-page", LinkTypePage},
			{"/sitemap.xml", LinkTypePage},
			{"/robots.txt", LinkTypePage},
			{"/assets/js/app.js", LinkTypeAsset},
			{"/assets/js/vendor.js", LinkTypeAsset},
		}

		for i, tt := range expectedLinks {
			link := result.PageData.Links[i]

			assert.Equal(t, server.URL+tt.Url, link.Url.String())
			assert.Equal(t, tt.LinkType, link.LinkType)
		}
	})
}

func Test_AnalizeUrlGet_NotFound(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/not-found")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusNotFound, result.HttpCode)
		assert.Equal(t, UrlTypeOther, result.UrlType)
		assert.Nil(t, result.PageData)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
	})
}

func Test_AnalizeUrlGet_Asset(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/assets/css/main.css")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypeOther, result.UrlType)
		assert.Nil(t, result.PageData)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
	})
}

func Test_AnalizeUrlGet_PageWithoutLinks(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/no-links")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypePage, result.UrlType)
		assert.Len(t, result.PageData.Links, 0)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
		assert.False(t, result.PageData.Seo.HasTitle)
		assert.Equal(t, "", result.PageData.Seo.Title)
		assert.False(t, result.PageData.Seo.HasDescription)
		assert.Equal(t, "", result.PageData.Seo.Description)
		assert.False(t, result.PageData.Seo.HasH1)
	})
}

func Test_AnalizeUrlGet_Sitemap(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/sitemap.xml")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypeOther, result.UrlType)
		assert.Nil(t, result.PageData)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
	})
}

func Test_AnalizeUrlGet_Error(t *testing.T) {
	httpClient := &http.Client{}

	parsedURL, err := url.Parse("http://127.0.0.1:19999/nonexistent")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	startTime := time.Now()
	result := analizeUrlGet(httpClient, parsedURL)
	endTime := time.Now()

	assert.Equal(t, parsedURL, result.Url)
	assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
	assert.NotNil(t, result.Error)
	assert.Equal(t, 0, result.HttpCode)
	assert.Equal(t, UrlTypeUnknown, result.UrlType)
	assert.Nil(t, result.PageData)
	assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
}

func Test_AnalizeUrlGet_DuplicateLinks(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/duplicate-links")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypePage, result.UrlType)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)

		expectedLinks := []struct {
			Url      string
			LinkType LinkType
		}{
			{"/assets/css/main.css", LinkTypeAsset},
			{"/", LinkTypePage},
			{"/about", LinkTypePage},
			{"/contact", LinkTypePage},
		}

		for i, tt := range expectedLinks {
			link := result.PageData.Links[i]

			assert.Equal(t, server.URL+tt.Url, link.Url.String())
			assert.Equal(t, tt.LinkType, link.LinkType)
		}
	})
}

func Test_AnalizeUrlGet_AnchorLinks(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/anchor-links")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypePage, result.UrlType)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)

		expectedLinks := []struct {
			Url      string
			LinkType LinkType
		}{
			{"/", LinkTypePage},
			{"/about", LinkTypePage},
		}

		for i, tt := range expectedLinks {
			link := result.PageData.Links[i]

			assert.Equal(t, server.URL+tt.Url, link.Url.String())
			assert.Equal(t, tt.LinkType, link.LinkType)
		}
	})
}

func Test_AnalizeUrlGet_ExternalLinks(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/external-links")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypePage, result.UrlType)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)

		expectedLinks := []struct {
			Url      string
			LinkType LinkType
		}{
			{server.URL + "/", LinkTypePage},
			{"https://www.google.com", LinkTypePage},
			{"https://www.github.com", LinkTypePage},
			{"https://www.example.com", LinkTypePage},
			{"http://external-site.com/page", LinkTypePage},
		}

		for i, tt := range expectedLinks {
			link := result.PageData.Links[i]

			assert.Equal(t, tt.Url, link.Url.String())
			assert.Equal(t, tt.LinkType, link.LinkType)
		}
	})
}

func Test_AnalizeUrlGet_MixedContent(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/mixed-content")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlGet(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodGet, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypePage, result.UrlType)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)

		expectedLinks := []struct {
			Url      string
			LinkType LinkType
		}{
			{server.URL + "/assets/css/main.css", LinkTypeAsset},
			{server.URL + "/", LinkTypePage},
			{server.URL + "/about", LinkTypePage},
			{server.URL + "/contact", LinkTypePage},
			{server.URL + "/blog", LinkTypePage},
			{"https://www.example.com", LinkTypePage},
			{server.URL + "/not-found", LinkTypePage},
			{server.URL + "/blog?page=2&sort=date", LinkTypePage},
			{server.URL + "/api/data?format=json", LinkTypePage},
			{server.URL + "/assets/images/logo.png", LinkTypeAsset},
			{server.URL + "/assets/images/banner.jpg", LinkTypeAsset},
			{"https://external.com/image.png", LinkTypeAsset},
			{server.URL + "/assets/js/app.js", LinkTypeAsset},
			{"https://external.com/script.js", LinkTypeAsset},
		}

		for i, tt := range expectedLinks {
			link := result.PageData.Links[i]

			assert.Equal(t, tt.Url, link.Url.String())
			assert.Equal(t, tt.LinkType, link.LinkType)
		}
	})
}

func Test_AnalizeUrlHead_Page(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlHead(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodHead, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypePage, result.UrlType)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
		assert.Nil(t, result.PageData)
	})
}

func Test_AnalizeUrlHead_NotFound(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/not-found")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlHead(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodHead, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusNotFound, result.HttpCode)
		assert.Equal(t, UrlTypeOther, result.UrlType)
		assert.Nil(t, result.PageData)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
	})
}

func Test_AnalizeUrlHead_Asset(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/assets/css/main.css")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlHead(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodHead, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypeOther, result.UrlType)
		assert.Nil(t, result.PageData)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
	})
}

func Test_AnalizeUrlHead_Sitemap(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient := http.DefaultClient

		parsedURL, err := url.Parse(server.URL + "/sitemap.xml")
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		startTime := time.Now()
		result := analizeUrlHead(httpClient, parsedURL)
		endTime := time.Now()

		assert.Equal(t, parsedURL, result.Url)
		assert.Equal(t, AnalizeMethodHead, result.AnalizeMethod)
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.HttpCode)
		assert.Equal(t, UrlTypeOther, result.UrlType)
		assert.Nil(t, result.PageData)
		assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
	})
}

func Test_AnalizeUrlHead_Error(t *testing.T) {
	httpClient := &http.Client{}

	parsedURL, err := url.Parse("http://127.0.0.1:19999/nonexistent")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	startTime := time.Now()
	result := analizeUrlHead(httpClient, parsedURL)
	endTime := time.Now()

	assert.Equal(t, parsedURL, result.Url)
	assert.Equal(t, AnalizeMethodHead, result.AnalizeMethod)
	assert.NotNil(t, result.Error)
	assert.Equal(t, 0, result.HttpCode)
	assert.Equal(t, UrlTypeUnknown, result.UrlType)
	assert.Nil(t, result.PageData)
	assert.WithinRange(t, result.DiscoveredAt, startTime, endTime)
}
