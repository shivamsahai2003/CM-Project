package services

import (
	"compress/gzip"
	"encoding/xml"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"adserving/models"
)

const yahooXMLURL = "https://contextual-stage.media.net/test/mock/provider/yahoo.xml"

var DefaultAds = []models.YahooAd{
	{
		TitleHTML: template.HTML("Shop Top Deals Today"),
		DescHTML:  template.HTML("Find amazing discounts on popular products. Limited time offers available now."),
		Link:      "https://example.com/deals",
		Host:      "example.com",
	},
	{
		TitleHTML: template.HTML("Compare Best Prices"),
		DescHTML:  template.HTML("Get the best prices from trusted retailers. Save money on your next purchase."),
		Link:      "https://example.com/compare",
		Host:      "example.com",
	},
	{
		TitleHTML: template.HTML("Exclusive Online Offers"),
		DescHTML:  template.HTML("Special offers only available online. Don't miss out on these savings."),
		Link:      "https://example.com/offers",
		Host:      "example.com",
	},
}

type YahooService struct {
	client *http.Client
}

func NewYahooService() *YahooService {
	return &YahooService{client: &http.Client{Timeout: 8 * time.Second}}
}

func (s *YahooService) FetchAds() ([]models.YahooAd, error) {
	req, err := http.NewRequest("GET", yahooXMLURL, nil)
	if err != nil {
		log.Printf("ads API request error: %v, using defaults", err)
		return DefaultAds, nil
	}
	req.Header.Set("User-Agent", "AdService/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("ads API fetch error: %v, using defaults", err)
		return DefaultAds, nil
	}
	defer resp.Body.Close()

	var reader io.ReadCloser = resp.Body
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		if gr, err := gzip.NewReader(resp.Body); err == nil {
			defer gr.Close()
			reader = gr
		}
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("ads API read error: %v, using defaults", err)
		return DefaultAds, nil
	}

	var doc models.YahooResults
	if err := xml.Unmarshal(raw, &doc); err != nil {
		log.Printf("ads API parse error: %v, using defaults", err)
		return DefaultAds, nil
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
		ads = append(ads, models.YahooAd{
			TitleHTML: template.HTML(html.UnescapeString(strings.TrimSpace(li.Title))),
			DescHTML:  template.HTML(html.UnescapeString(strings.TrimSpace(li.Description))),
			Link:      link,
			Host:      strings.TrimSpace(li.SiteHost),
		})
	}

	if len(ads) == 0 {
		log.Printf("ads API returned no ads, using defaults")
		return DefaultAds, nil
	}

	return ads, nil
}
