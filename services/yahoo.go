package services

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"adserving/models"
)

const yahooXMLURL = "https://contextual-stage.media.net/test/mock/provider/yahoo.xml"

// YahooService handles Yahoo XML ad fetching
type YahooService struct {
	client *http.Client
}

// NewYahooService creates a new Yahoo service
func NewYahooService() *YahooService {
	return &YahooService{
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

// FetchAds fetches and parses Yahoo XML ads
func (s *YahooService) FetchAds() ([]models.YahooAd, error) {
	req, _ := http.NewRequest("GET", yahooXMLURL, nil)
	req.Header.Set("User-Agent", "KeywordService/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch yahoo xml: %w", err)
	}
	defer resp.Body.Close()

	var reader io.ReadCloser = resp.Body
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		gzr, gzErr := gzip.NewReader(resp.Body)
		if gzErr == nil {
			defer gzr.Close()
			reader = gzr
		}
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read yahoo xml: %w", err)
	}

	var doc models.YahooResults
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("xml unmarshal: %w", err)
	}

	var ads []models.YahooAd
	for _, li := range doc.ResultSet.Listings {
		link := strings.TrimSpace(li.ClickUrl.URL)
		if link == "" && len(li.Extensions.ActionExtension.Items) > 0 {
			link = strings.TrimSpace(li.Extensions.ActionExtension.Items[0].Link)
		}
		if link == "" {
			continue
		}
		title := template.HTML(html.UnescapeString(strings.TrimSpace(li.Title)))
		desc := template.HTML(html.UnescapeString(strings.TrimSpace(li.Description)))
		ads = append(ads, models.YahooAd{
			TitleHTML: title,
			DescHTML:  desc,
			Link:      link,
			Host:      strings.TrimSpace(li.SiteHost),
		})
	}

	return ads, nil
}
