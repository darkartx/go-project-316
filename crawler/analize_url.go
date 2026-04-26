package crawler

import (
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type UrlType uint8

const (
	UrlTypeUnknown UrlType = iota
	UrlTypeOther
	UrlTypePage
)

type AnalizeMethod uint8

const (
	AnalizeMethodHead AnalizeMethod = iota
	AnalizeMethodGet
)

type AnalizeUrlResult struct {
	Url           *url.URL
	AnalizeMethod AnalizeMethod
	Error         error
	HttpCode      int
	UrlType       UrlType
	PageData      *PageData
	DiscoveredAt  time.Time
}

type PageData struct {
	Links []PageDataLink
}

type LinkType uint8

const (
	LinkTypeAsset = iota
	LinkTypePage
)

type PageDataLink struct {
	Url      *url.URL
	LinkType LinkType
}

func analizeUrlGet(httpClient *http.Client, url *url.URL) *AnalizeUrlResult {
	var result AnalizeUrlResult

	result.Url = url
	result.AnalizeMethod = AnalizeMethodGet

	res, err := httpClient.Get(url.String())

	if err != nil {
		result.Error = err
	} else {
		defer res.Body.Close()

		result.HttpCode = res.StatusCode
		result.UrlType = detectUrlType(res)

		if result.UrlType == UrlTypePage {
			if pageData, err := extractPageData(res.Body, url); err != nil {
				result.Error = err
			} else {
				result.PageData = pageData
			}
		}
	}

	result.DiscoveredAt = time.Now()

	return &result
}

func analizeUrlHead(httpClient *http.Client, url *url.URL) *AnalizeUrlResult {
	var result AnalizeUrlResult

	result.Url = url
	result.AnalizeMethod = AnalizeMethodHead

	res, err := httpClient.Head(url.String())

	if err != nil {
		result.Error = err
	} else {
		result.HttpCode = res.StatusCode
		result.UrlType = detectUrlType(res)
	}

	result.DiscoveredAt = time.Now()

	return &result
}

func detectUrlType(response *http.Response) UrlType {
	contentType := response.Header.Get("Content-Type")

	if contentType == "" {
		return UrlTypeUnknown
	}

	mediaType, _, err := mime.ParseMediaType(contentType)

	if err != nil {
		return UrlTypeUnknown
	} else {
		return mediaTypeToUrlType(mediaType)
	}
}

func mediaTypeToUrlType(mediaType string) UrlType {
	switch mediaType {
	case "text/html":
		return UrlTypePage
	default:
		return UrlTypeOther
	}
}

func extractPageData(body io.Reader, rootUrl *url.URL) (*PageData, error) {
	links := make([]PageDataLink, 0)
	result := &PageData{links}

	seenLinks := make(map[string]struct{})

	tokenizer := html.NewTokenizer(body)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		token := tokenizer.Token()
		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}

		switch token.Data {
		case "a", "link", "script", "img", "source", "iframe":
			link := extractHtmlLink(token, rootUrl)
			if link != nil {
				linkUrl := link.Url.String()
				if _, ok := seenLinks[linkUrl]; !ok {
					seenLinks[linkUrl] = struct{}{}
					result.Links = append(result.Links, *link)
				}
			}
		default:
			continue
		}

	}

	return result, nil
}

func extractHtmlLink(token html.Token, rootUrl *url.URL) *PageDataLink {
	var attr string
	var linkType LinkType

	switch token.Data {
	case "a":
		attr = "href"
		linkType = LinkTypePage
	case "link":
		attr = "href"
		linkType = LinkTypeAsset
	case "script", "img", "source":
		attr = "src"
		linkType = LinkTypeAsset
	case "iframe":
		attr = "src"
		linkType = LinkTypePage
	}

	for _, a := range token.Attr {
		if a.Key != attr {
			continue
		}

		href := strings.TrimSpace(a.Val)
		if href == "" || strings.HasPrefix(href, "#") {
			return nil
		}

		parsed, err := url.Parse(href)
		if err != nil {
			return nil
		}

		resolved := rootUrl.ResolveReference(parsed)
		resolved.Fragment = ""

		switch resolved.Scheme {
		case "http", "https":
		default:
			continue
		}

		return &PageDataLink{resolved, linkType}
	}

	return nil
}
