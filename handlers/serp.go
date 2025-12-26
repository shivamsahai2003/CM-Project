package handlers

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"adserving/config"
	"adserving/db"
	"adserving/models"
	"adserving/services"
	"adserving/utils"
)

const dummySerpTemplate = "storage/html/SerpTemplateDummy.html"

type SerpHandler struct {
	yahooService *services.YahooService
}

func NewSerpHandler(yahooService *services.YahooService) *SerpHandler {
	return &SerpHandler{yahooService: yahooService}
}

func (h *SerpHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userAgent := r.UserAgent()
	isBot := utils.IsBotUA(userAgent)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	q := r.URL.Query()
	params := models.SerpParams{
		Query:       q.Get("q"),
		Slot:        q.Get("slot"),
		CountryCode: q.Get("cc"),
		KeywordID:   q.Get("kid"),
		PublisherID: q.Get("pid"),
	}

	clientIP := utils.GetClientIP(r)
	publisherID := utils.AtoiOrZero(params.PublisherID)
	keywordID := utils.AtoiOrZero(params.KeywordID)

	rule := config.GetRuleByPublisherIDAndUserAgent(publisherID, userAgent)

	if rule.Action.Block {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "<h2>403 – Traffic blocked</h2>")
		return
	}

	// Record keyword click (ignore DB errors)
	if publisherID > 0 && db.GetDB() != nil {
		db.GetDB().Exec(
			`INSERT INTO keyword_click (publisher_id, keyword_id, keyword_title, slot, client_ip, user_agent, country_code) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			publisherID, keywordID, params.Query, params.Slot, clientIP, userAgent, params.CountryCode,
		)
	}

	// Get template path, fallback to dummy
	serpTemplatePath := "storage/html/" + rule.Action.SerpTemplateID
	maxAds := utils.CountAdSlots(serpTemplatePath)
	if maxAds == 0 {
		serpTemplatePath = dummySerpTemplate
		maxAds = 3
	}

	var ads []models.YahooAd
	if !isBot {
		// FetchAds returns defaults on error
		ads, _ = h.yahooService.FetchAds()

		if len(ads) > maxAds {
			ads = ads[:maxAds]
		}

		// Record ad impressions (ignore DB errors)
		if db.GetDB() != nil {
			for pos, ad := range ads {
				db.GetDB().Exec(
					`INSERT INTO ad_impression (publisher_id, keyword_id, keyword_title, ad_position, ad_title, ad_host, client_ip, user_agent, country_code) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
					publisherID, keywordID, params.Query, pos+1, string(ad.TitleHTML), ad.Host, clientIP, userAgent, params.CountryCode,
				)
			}
		}
	}

	title := "SERP"
	if params.Query != "" {
		title = "Results for: " + params.Query
	}

	var adsVM []models.AdViewModel
	for _, ad := range ads {
		qs := url.Values{}
		qs.Set("u", ad.Link)
		qs.Set("slot", params.Slot)
		qs.Set("kid", params.KeywordID)
		qs.Set("q", params.Query)
		qs.Set("adhost", ad.Host)
		qs.Set("adtitle", string(ad.TitleHTML))
		qs.Set("pid", params.PublisherID)
		qs.Set("cc", params.CountryCode)

		adsVM = append(adsVM, models.AdViewModel{
			TitleHTML:   ad.TitleHTML,
			DescHTML:    ad.DescHTML,
			Host:        ad.Host,
			ClickHref:   "/ad-click?" + qs.Encode(),
			RenderLinks: !isBot,
		})
	}

	dataMap := map[string]interface{}{
		"Title":  html.EscapeString(title),
		"IsBot":  isBot,
		"HasAds": len(adsVM) > 0,
		"Ads":    adsVM,
	}
	for i, ad := range adsVM {
		idx := strconv.Itoa(i + 1)
		dataMap["AdTitle"+idx] = ad.TitleHTML
		dataMap["AdDesc"+idx] = ad.DescHTML
		dataMap["AdHref"+idx] = ad.ClickHref
	}

	tmpl, err := template.ParseFiles(serpTemplatePath)
	if err != nil {
		log.Printf("template parse error: %v, trying dummy", err)
		tmpl, err = template.ParseFiles(dummySerpTemplate)
		if err != nil {
			fmt.Fprint(w, "<h2>Template error</h2>")
			return
		}
	}
	if err := tmpl.Execute(w, dataMap); err != nil {
		log.Printf("template execute error: %v", err)
	}
}
