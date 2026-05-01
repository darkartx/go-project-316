package crawler

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type AnalizeUrl struct {
	Url    *url.URL
	Result *AnalizeUrlResult
	Method AnalizeMethod
}

func (a *AnalizeUrl) analize(httpClient *http.Client) {
	switch a.Method {
	case AnalizeMethodGet:
		a.Result = analizeUrlGet(httpClient, a.Url)
	case AnalizeMethodHead:
		a.Result = analizeUrlHead(httpClient, a.Url)
	}
}

type UrlState struct {
	Depth    uint
	LinkType LinkType
}

type Analizer struct {
	httpClient *http.Client
	doneUrls   map[string]*AnalizeUrl
	inProcess  map[string]struct{}
	states     map[string]*UrlState
	inCh       chan *AnalizeUrl
	doneCh     chan *AnalizeUrl
	stopCh     chan struct{}
	rootUrl    *url.URL
	depth      uint
}

func NewAnalizer(rootUrl *url.URL, depth uint, httpClient *http.Client) *Analizer {
	doneUrls := make(map[string]*AnalizeUrl)
	inProcess := make(map[string]struct{})
	states := make(map[string]*UrlState)

	return &Analizer{
		httpClient: httpClient,
		doneUrls:   doneUrls,
		inProcess:  inProcess,
		states:     states,
		rootUrl:    rootUrl,
		depth:      depth,
	}
}

func (a *Analizer) Analize(ctx context.Context) (Report, error) {
	a.inCh = make(chan *AnalizeUrl)
	a.doneCh = make(chan *AnalizeUrl)
	a.stopCh = make(chan struct{})

	go processWorker(a.httpClient, a.inCh, a.doneCh, a.stopCh)

	a.analizeUrl(a.rootUrl, AnalizeMethodGet, 0, LinkTypePage)
	a.handleResult(ctx)

	return a.Report(), nil
}

func (a *Analizer) handleResult(ctx context.Context) {
	for {
		select {
		case result := <-a.doneCh:
			a.postProcess(result)

			if len(a.inProcess) == 0 {
				close(a.stopCh)
				return
			}
		case <-ctx.Done():
			close(a.stopCh)
			return
		}
	}
}

func (a *Analizer) analizeUrl(url *url.URL, method AnalizeMethod, depth uint, linkType LinkType) bool {
	urlStr := url.String()

	if urlState, exists := a.states[urlStr]; exists {
		if urlState.LinkType < linkType {
			urlState.LinkType = linkType
		}

		if urlState.Depth > depth {
			urlState.Depth = depth

			if analizeUrl, exists := a.doneUrls[urlStr]; exists {
				a.processNextUrls(analizeUrl)
			}
		}

		return false
	}

	a.states[urlStr] = &UrlState{
		Depth:    depth,
		LinkType: linkType,
	}

	analizeUrl := &AnalizeUrl{
		Url:    url,
		Method: method,
	}

	a.inProcess[urlStr] = struct{}{}

	select {
	case a.inCh <- analizeUrl:
		return true
	case <-a.stopCh:
		return false
	}
}

func (a *Analizer) postProcess(analizeUrl *AnalizeUrl) {
	urlStr := analizeUrl.Url.String()
	delete(a.inProcess, urlStr)

	a.processNextUrls(analizeUrl)

	a.doneUrls[urlStr] = analizeUrl
}

func (a *Analizer) processNextUrls(analizeUrl *AnalizeUrl) {
	urlStr := analizeUrl.Url.String()
	var depth uint = 0

	if urlState, exists := a.states[urlStr]; exists {
		depth = urlState.Depth + 1
	}

	if depth > a.depth ||
		analizeUrl.Result.UrlType != UrlTypePage ||
		analizeUrl.Result.PageData == nil ||
		!isRootUrlEquals(analizeUrl.Result.Url, a.rootUrl) {
		return
	}

	for _, link := range analizeUrl.Result.PageData.Links {
		a.analizeUrl(link.Url, analizeMethod(link), depth, link.LinkType)
	}
}

func (a *Analizer) Report() Report {
	var report Report

	report.RootURL = a.rootUrl.String()
	report.Depth = a.depth
	report.GeneratedAt = time.Now()
	report.Pages = make([]ReportPage, 0)

	for _, page := range a.doneUrls {
		urlStr := page.Url.String()
		urlState := a.states[urlStr]

		if urlState.Depth > a.depth ||
			urlState.LinkType != LinkTypePage ||
			!isRootUrlEquals(page.Url, a.rootUrl) {
			continue
		}

		report.Pages = append(report.Pages, a.makePageReport(page, urlState.Depth))
	}

	return report
}

func (a *Analizer) makePageReport(page *AnalizeUrl, depth uint) ReportPage {
	var reportPage ReportPage

	reportPage.URL = page.Result.Url.String()
	reportPage.Depth = depth
	reportPage.DiscoveredAt = page.Result.DiscoveredAt
	reportPage.HTTPStatus = page.Result.HttpCode
	reportPage.Status = "ok"

	if page.Result.Error != nil {
		reportPage.Error = page.Result.Error.Error()
	}

	reportPage.BrokenLinks = make([]ReportPageBrokenLink, 0)
	if page.Result.PageData != nil {
		reportPage.Seo.HasTitle = page.Result.PageData.Seo.HasTitle
		reportPage.Seo.Title = page.Result.PageData.Seo.Title
		reportPage.Seo.HasDescription = page.Result.PageData.Seo.HasDescription
		reportPage.Seo.Description = page.Result.PageData.Seo.Description
		reportPage.Seo.HasH1 = page.Result.PageData.Seo.HasH1

		for _, link := range page.Result.PageData.Links {
			item := a.doneUrls[link.Url.String()]

			if item == nil {
				continue
			}

			if item.Result != nil && (item.Result.Error != nil || item.Result.HttpCode >= 400) {
				var errorString string
				if item.Result.Error != nil {
					errorString = item.Result.Error.Error()
				}
				brokenLink := ReportPageBrokenLink{item.Url.String(), item.Result.HttpCode, errorString}
				reportPage.BrokenLinks = append(reportPage.BrokenLinks, brokenLink)
			}
		}
	}

	return reportPage
}

func processWorker(httpClient *http.Client, inCh chan *AnalizeUrl, doneCh chan *AnalizeUrl, stopCh chan struct{}) {
	for {
		select {
		case analizeUrl := <-inCh:
			go func(analizeUrl *AnalizeUrl) {
				analizeUrl.analize(httpClient)

				select {
				case <-stopCh:
				case doneCh <- analizeUrl:
				}
			}(analizeUrl)
		case <-stopCh:
			return
		}
	}
}

func isRootUrlEquals(left *url.URL, right *url.URL) bool {
	return left.Host == right.Host
}

func analizeMethod(link PageDataLink) AnalizeMethod {
	switch link.LinkType {
	case LinkTypePage:
		return AnalizeMethodGet
	default:
		return AnalizeMethodHead
	}
}
