package crawler

import (
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type UrlType uint8

const (
	UrlTypeOther UrlType = iota
	UrlTypeImage
	UrlTypeScript
	UrlTypeStyle
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
	ContentLength uint
	DiscoveredAt  time.Time
}

type PageData struct {
	Links []PageDataLink
	Seo   PageDataSeo
}

type LinkType uint8

const (
	LinkTypeOther = iota
	LinkTypeStyle
	LinkTypeImage
	LinkTypeScript
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
		result.ContentLength = extractContentLength(res)

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
		result.ContentLength = extractContentLength(res)
	}

	result.DiscoveredAt = time.Now()

	return &result
}

func detectUrlType(response *http.Response) UrlType {
	contentType := response.Header.Get("Content-Type")

	if contentType == "" {
		return UrlTypeOther
	}

	mediaType, _, err := mime.ParseMediaType(contentType)

	if err != nil {
		return UrlTypeOther
	} else {
		return mediaTypeToUrlType(mediaType)
	}
}

func mediaTypeToUrlType(mediaType string) UrlType {
	switch mediaType {
	case "text/html":
		return UrlTypePage
	case "text/css":
		return UrlTypeStyle
	case "application/javascript":
		return UrlTypeScript
	default:
		if strings.HasPrefix(mediaType, "image") {
			return UrlTypeImage
		}
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
		var link *PageDataLink

		switch token.Data {
		case "a":
			link = extractALink(token, rootUrl)
		case "iframe":
			link = extractIframeLink(token, rootUrl)
		case "link":
			link = extractLinkLink(token, rootUrl)
		case "script":
			link = extractScriptLink(token, rootUrl)
		case "img":
			link = extractImgLink(token, rootUrl)
		case "source":
			link = extractSourceLink(token, rootUrl)
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
		}

		if link != nil {
			linkUrl := link.Url.String()
			if _, ok := seenLinks[linkUrl]; !ok {
				seenLinks[linkUrl] = struct{}{}
				result.Links = append(result.Links, *link)
			}
		}
	}

	return result, nil
}

func extractALink(token html.Token, rootUrl *url.URL) *PageDataLink {
	for _, a := range token.Attr {
		if a.Key != "href" {
			continue
		}

		resolved := resolveLink(a.Val, rootUrl)
		if resolved == nil {
			return nil
		}

		return &PageDataLink{resolved, LinkTypePage}
	}

	return nil
}

func extractLinkLink(token html.Token, rootUrl *url.URL) *PageDataLink {
	var rel string

	for _, a := range token.Attr {
		if a.Key != "rel" {
			continue
		}

		rel = strings.TrimSpace(a.Val)
		break
	}

	for _, a := range token.Attr {
		if a.Key != "href" {
			continue
		}

		resolved := resolveLink(a.Val, rootUrl)
		if resolved == nil {
			return nil
		}

		var linkType LinkType

		switch rel {
		case "stylesheet":
			linkType = LinkTypeStyle
		case "icon":
			linkType = LinkTypeImage
		default:
			linkType = LinkTypeOther
		}

		return &PageDataLink{resolved, linkType}
	}

	return nil
}

func extractScriptLink(token html.Token, rootUrl *url.URL) *PageDataLink {
	for _, a := range token.Attr {
		if a.Key != "src" {
			continue
		}

		resolved := resolveLink(a.Val, rootUrl)
		if resolved == nil {
			return nil
		}

		return &PageDataLink{resolved, LinkTypeScript}
	}

	return nil
}

func extractIframeLink(token html.Token, rootUrl *url.URL) *PageDataLink {
	for _, a := range token.Attr {
		if a.Key != "src" {
			continue
		}

		resolved := resolveLink(a.Val, rootUrl)
		if resolved == nil {
			return nil
		}

		return &PageDataLink{resolved, LinkTypePage}
	}

	return nil
}

func extractImgLink(token html.Token, rootUrl *url.URL) *PageDataLink {
	for _, a := range token.Attr {
		if a.Key != "src" {
			continue
		}

		resolved := resolveLink(a.Val, rootUrl)
		if resolved == nil {
			return nil
		}

		return &PageDataLink{resolved, LinkTypeImage}
	}

	return nil
}

func extractSourceLink(token html.Token, rootUrl *url.URL) *PageDataLink {
	var typ string

	for _, a := range token.Attr {
		if a.Key != "type" {
			continue
		}

		typ = strings.TrimSpace(a.Val)
		break
	}

	for _, a := range token.Attr {
		if a.Key != "src" {
			continue
		}

		resolved := resolveLink(a.Val, rootUrl)
		if resolved == nil {
			return nil
		}

		var linkType LinkType

		if strings.HasPrefix(typ, "image") {
			linkType = LinkTypeImage
		} else {
			linkType = LinkTypeOther
		}

		return &PageDataLink{resolved, linkType}
	}

	return nil
}

func resolveLink(link string, rootUrl *url.URL) *url.URL {
	link = strings.TrimSpace(link)
	if link == "" || strings.HasPrefix(link, "#") {
		return nil
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return nil
	}

	resolved := rootUrl.ResolveReference(parsed)
	resolved.Fragment = ""

	switch resolved.Scheme {
	case "http", "https":
		return resolved
	default:
		return nil
	}
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

func extractContentLength(response *http.Response) uint {
	contentLength := response.Header.Get("Content-Length")

	if contentLength == "" {
		return 0
	}

	result, err := strconv.ParseUint(contentLength, 10, 32)
	if err != nil {
		return 0
	}

	return uint(result)
}
