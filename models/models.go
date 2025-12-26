package models

import (
	"encoding/json"
	"html/template"
)

type KeywordItem struct {
	Title string      `json:"t"`
	ID    json.Number `json:"i"`
}

type KeywordResponse struct {
	Keywords []KeywordItem `json:"k"`
}

type YahooResults struct {
	ResultSet struct {
		Listings []YahooListing `xml:"Listing"`
	} `xml:"ResultSet"`
}

type YahooListing struct {
	Rank        string     `xml:"rank,attr"`
	Title       string     `xml:"title,attr"`
	Description string     `xml:"description,attr"`
	SiteHost    string     `xml:"siteHost,attr"`
	ClickUrl    YahooClick `xml:"ClickUrl"`
	Extensions  YahooExt   `xml:"Extensions"`
}

type YahooExt struct {
	ActionExtension struct {
		Items []struct {
			Text string `xml:"text"`
			Link string `xml:"link"`
		} `xml:"actionItem"`
	} `xml:"actionExtension"`
}

type YahooClick struct {
	Type string `xml:"type,attr"`
	URL  string `xml:",chardata"`
}

type YahooAd struct {
	TitleHTML template.HTML
	DescHTML  template.HTML
	Link      string
	Host      string
}

type ClickStatKey struct {
	Slot      string
	KeywordID string
	Query     string
	AdHost    string
}

type RenderParams struct {
	Slot         string
	CountryCode  string
	TemplateSize string
	PublisherID  string
	Domain       string
	LayoutID     string
	MaxNumber    int
	PageTitle    string
	ReferrerURL  string
	KeywordRef   string
}

type SerpParams struct {
	Query       string
	Slot        string
	CountryCode string
	KeywordID   string
	PublisherID string
}

type AdViewModel struct {
	TitleHTML   template.HTML
	DescHTML    template.HTML
	Host        string
	ClickHref   string
	RenderLinks bool
}
