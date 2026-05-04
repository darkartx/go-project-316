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
	Seo   PageDataSeo
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

type PageDataSeo struct {
	HasTitle       bool
	Title          string
	HasDescription bool
	Description    string
	HasH1          bool
}

func analizeUrlGet(httpClient *HTTPClient, url *url.URL) *AnalizeUrlResult {
	var result AnalizeUrlResult

	result.Url = url
	result.AnalizeMethod = AnalizeMethodGet

	res, err := httpClient.Get(url.String())

	if err != nil {
		result.Error = err
	} else {
		defer res.Body.Close() //nolint:errcheck

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

func analizeUrlHead(httpClient *HTTPClient, url *url.URL) *AnalizeUrlResult {
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
	result := &PageData{links, PageDataSeo{}}

	seenLinks := make(map[string]struct{})

	tokenizer := html.NewTokenizer(body)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}

		token := tokenizer.Token()

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
		case "title":
			result.Seo.HasTitle = true
			result.Seo.Title = extractHtmlString(token, tokenizer)
		case "meta":
			value, is := extractDescription(token)

			if is {
				result.Seo.HasDescription = true
				result.Seo.Description = value
			}
		case "h1":
			result.Seo.HasH1 = true
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

func extractHtmlString(startToken html.Token, tokenizer *html.Tokenizer) string {
	var sb strings.Builder

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		token := tokenizer.Token()
		if tt == html.EndTagToken && token.Data == startToken.Data {
			break
		}

		sb.WriteString(token.String())
	}

	return html.UnescapeString(sb.String())
}

func extractDescription(token html.Token) (string, bool) {
	var result string
	var isDescription bool

	for _, a := range token.Attr {
		if a.Key == "name" && a.Val == "description" {
			isDescription = true
		}

		if a.Key == "content" {
			result = a.Val
		}
	}

	if isDescription {
		return html.UnescapeString(result), true
	}

	return "", false
}
