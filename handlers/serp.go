package handlers

import (
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"adserving/config"
	"adserving/db"
	"adserving/models"
	"adserving/services"
	"adserving/templates"
	"adserving/utils"
)

// SerpHandler handles SERP page requests
type SerpHandler struct {
	yahooService *services.YahooService
}

// NewSerpHandler creates a new SERP handler
func NewSerpHandler(yahooService *services.YahooService) *SerpHandler {
	return &SerpHandler{
		yahooService: yahooService,
	}
}

// Handle processes SERP page requests
func (h *SerpHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()
	isBot := utils.IsBotUA(ua)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	params := models.SerpParams{
		Q:      r.URL.Query().Get("q"),
		Slot:   r.URL.Query().Get("slot"),
		CC:     r.URL.Query().Get("cc"),
		D:      r.URL.Query().Get("d"),
		RURL:   r.URL.Query().Get("rurl"),
		PTitle: r.URL.Query().Get("ptitle"),
		LID:    r.URL.Query().Get("lid"),
		TSize:  r.URL.Query().Get("tsize"),
		KwRf:   r.URL.Query().Get("kwrf"),
		KID:    r.URL.Query().Get("kid"),
		PID:    r.URL.Query().Get("pid"),
		MaxAds: r.URL.Query().Get("maxads"),
	}

	// Log keyword click
	clientID := utils.GetClientIP(r)
	pubID := utils.AtoiOrZero(params.LID)
	kidInt := utils.AtoiOrZero(params.KID)

	var slotSQL any = nil
	if s := strings.TrimSpace(params.Slot); s != "" {
		if sInt, err := strconv.Atoi(s); err == nil {
			slotSQL = sInt
		}
	}

	if pubID > 0 {
		_, err := db.GetDB().Exec(
			"INSERT INTO keyword_click (slot_id, kid, `time`, `user id`, keyword_title, publisher_id, user_agent) VALUES (?, ?, NOW(), ?, ?, ?, ?)",
			slotSQL, kidInt, clientID, params.Q, pubID, ua,
		)
		if err != nil {
			log.Printf("insert keyword_click error: %v", err)
		}
	}

	ads, err := h.yahooService.FetchAds()
	if err != nil {
		log.Printf("Yahoo XML fetch/parse error: %v", err)
	}

	// Limit ads based on maxads param (passed from render.js)
	maxAds := config.DefaultPublisherConfig.MaxAds // Default fallback
	if params.MaxAds != "" {
		if n, err := strconv.Atoi(params.MaxAds); err == nil && n > 0 {
			maxAds = n
		}
	}
	log.Printf("SERP: Domain=%s, MaxAds param=%s, limiting to %d ads", params.D, params.MaxAds, maxAds)
	if len(ads) > maxAds {
		ads = ads[:maxAds]
	}

	title := "SERP"
	if params.Q != "" {
		title = "Results for: " + params.Q
	}

	// Build ad-click URLs; include lid so /ad-click can log publisher_id
	var adsVM []models.AdViewModel
	for _, a := range ads {
		qs := url.Values{}
		qs.Set("u", a.Link)
		if params.Slot != "" {
			qs.Set("slot", params.Slot)
		}
		if params.KID != "" {
			qs.Set("kid", params.KID)
		}
		if params.Q != "" {
			qs.Set("q", params.Q)
		}
		if a.Host != "" {
			qs.Set("adhost", a.Host)
		}
		if params.LID != "" {
			qs.Set("lid", params.LID)
		}
		clickHref := "/ad-click?" + qs.Encode()
		adsVM = append(adsVM, models.AdViewModel{
			TitleHTML:   a.TitleHTML,
			DescHTML:    a.DescHTML,
			Host:        a.Host,
			ClickHref:   clickHref,
			RenderLinks: !isBot,
		})
	}

	data := models.SerpPageData{
		Title:  html.EscapeString(title),
		Slot:   html.EscapeString(params.Slot),
		CC:     html.EscapeString(params.CC),
		D:      html.EscapeString(params.D),
		RURL:   html.EscapeString(params.RURL),
		PTitle: html.EscapeString(params.PTitle),
		LID:    html.EscapeString(params.LID),
		TSize:  html.EscapeString(params.TSize),
		KwRf:   html.EscapeString(params.KwRf),
		KID:    html.EscapeString(params.KID),
		PID:    html.EscapeString(params.PID),
		IsBot:  isBot,
		HasAds: len(adsVM) > 0,
		Ads:    adsVM,
	}

	t, _ := template.New("serp").Parse(templates.SerpTemplate)
	if err := t.Execute(w, data); err != nil {
		log.Printf("template execute error: %v", err)
	}
}
