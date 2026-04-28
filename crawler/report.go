package crawler

import "time"

type Report struct {
	RootURL     string       `json:"root_url"`
	Depth       uint         `json:"depth"`
	GeneratedAt time.Time    `json:"generated_at"`
	Pages       []ReportPage `json:"pages"`
}

type ReportPage struct {
	URL          string                 `json:"url"`
	Depth        uint                   `json:"depth"`
	HTTPStatus   int                    `json:"http_status"`
	Status       string                 `json:"status"`
	Seo          ReportPageSeo          `json:"seo"`
	Error        string                 `json:"error"`
	BrokenLinks  []ReportPageBrokenLink `json:"broken_links"`
	DiscoveredAt time.Time              `json:"discovered_at"`
}

type ReportPageBrokenLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

type ReportPageSeo struct {
	HasTitle       bool   `json:"has_title"`
	Title          string `json:"title,omitempty"`
	HasDescription bool   `json:"has_description"`
	Description    string `json:"description,omitempty"`
	HasH1          bool   `json:"has_h1"`
}
