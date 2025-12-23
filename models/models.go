package models

import (
	"encoding/json"
	"html/template"
)

// KeywordItem represents a keyword from the API response
type KeywordItem struct {
	T  string      `json:"t"`
	DT string      `json:"dt"`
	I  json.Number `json:"i"`
}

// KeywordResponse represents the keyword API response
type KeywordResponse struct {
	K []KeywordItem `json:"k"`
}

// YahooResults represents the Yahoo XML response structure
type YahooResults struct {
	ResultSet struct {
		Listings []YahooListing `xml:"Listing"`
	} `xml:"ResultSet"`
}

// YahooListing represents a single listing in Yahoo results
type YahooListing struct {
	Rank        string     `xml:"rank,attr"`
	Title       string     `xml:"title,attr"`
	Description string     `xml:"description,attr"`
	SiteHost    string     `xml:"siteHost,attr"`
	ClickUrl    YahooClick `xml:"ClickUrl"`
	Extensions  YahooExt   `xml:"Extensions"`
}

// YahooClick represents the click URL in a Yahoo listing
type YahooClick struct {
	Type string `xml:"type,attr"`
	URL  string `xml:",chardata"`
}

// YahooExt represents extensions in a Yahoo listing
type YahooExt struct {
	ActionExtension struct {
		Items []struct {
			Text string `xml:"text"`
			Link string `xml:"link"`
		} `xml:"actionItem"`
	} `xml:"actionExtension"`
}

// YahooAd represents a parsed Yahoo ad
type YahooAd struct {
	TitleHTML template.HTML
	DescHTML  template.HTML
	Link      string
	Host      string
}

// ClickStatKey is the key for click statistics
type ClickStatKey struct {
	Slot   string
	KID    string
	Q      string
	AdHost string
}

// RenderParams holds parameters for the render.js handler
type RenderParams struct {
	Slot   string
	Maxno  string
	CC     string
	LID    string
	D      string
	RURL   string
	PTitle string
	TSize  string
	KwRf   string
	PID    string
}

// SerpParams holds parameters for the SERP handler
type SerpParams struct {
	Q      string
	Slot   string
	CC     string
	D      string
	RURL   string
	PTitle string
	LID    string
	TSize  string
	KwRf   string
	KID    string
	PID    string
	MaxAds string
}

// AdViewModel represents an ad for rendering in templates
type AdViewModel struct {
	TitleHTML   template.HTML
	DescHTML    template.HTML
	Host        string
	ClickHref   string
	RenderLinks bool
}

//SerpPageData holds data for the SERP page template
// type SerpPageData struct {
// 	Title  string
// 	Slot   string
// 	CC     string
// 	D      string
// 	RURL   string
// 	PTitle string
// 	LID    string
// 	TSize  string
// 	KwRf   string
// 	KID    string
// 	PID    string
// 	IsBot  bool
// 	HasAds bool
// 	//Ads    []AdViewModel
// 	//Serp template required field
// 	AdTitle1 template.HTML
// 	AdDesc1  template.HTML
// 	AdHref1  string
// 	AdTitle2 template.HTML
// 	AdDesc2  template.HTML
// 	AdHref2  string
// 	AdTitle3 template.HTML
// 	AdDesc3  template.HTML
// 	AdHref3  string
// }

// keyword struct for placeholder
type Keyword struct {
	Title string
	Href  string
}
type KeywordSlots struct {
	PID      int
	Keywords []Keyword
}
