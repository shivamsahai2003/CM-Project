package handlers

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"adserving/config"
	"adserving/db"
	"adserving/models"
	"adserving/services"
	"adserving/utils"
)

// SerpHandler handles SERP page requests
func CountAdPlaceHolders(templateStr string) int {
	re := regexp.MustCompile(`\{\{\.ad_desc_\d+\}\}`)
	matches := re.FindAllString(templateStr, -1)
	return len(matches)
}

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
	log.Printf("result of params: %+v", params)
	// Log keyword click
	clientID := utils.GetClientIP(r) // todo recheck
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

	// Get publisher config by PID
	pid := utils.AtoiOrZero(params.PID)
	pubConfig := config.GetPublisherConfigByPID(pid)

	// Get max ads from SERP template (template decides how many ads to show)
	serpTemplatePath := "storage/html/" + pubConfig.SerpTemplate
	maxAds := utils.CountAdSlots(serpTemplatePath)
	if maxAds == 0 {
		maxAds = 3 // fallback
	}

	// Override with maxads param if provided (from render.js)
	if params.MaxAds != "" {
		if n, err := strconv.Atoi(params.MaxAds); err == nil && n > 0 && n < maxAds {
			maxAds = n
		}
	}
	fmt.Println("ads for serp", ads)
	log.Printf("SERP: PID=%d, Domain=%s, SerpTemplate=%s, MaxAds=%d", pid, params.D, pubConfig.SerpTemplate, maxAds)
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

	dataMap := map[string]interface{}{
		"Title":  html.EscapeString(title),
		"Slot":   html.EscapeString(params.Slot),
		"CC":     html.EscapeString(params.CC),
		"D":      html.EscapeString(params.D),
		"RURL":   html.EscapeString(params.RURL),
		"PTitle": html.EscapeString(params.PTitle),
		"LID":    html.EscapeString(params.LID),
		"TSize":  html.EscapeString(params.TSize),
		"KwRf":   html.EscapeString(params.KwRf),
		"KID":    html.EscapeString(params.KID),
		"PID":    html.EscapeString(params.PID),
		"IsBot":  isBot,
		"HasAds": len(adsVM) > 0,
		"Ads":    adsVM, //added this
	}
	for index, ad := range adsVM {
		n := strconv.Itoa(index + 1)
		dataMap["AdTitle"+n] = ad.TitleHTML
		dataMap["AdDesc"+n] = ad.DescHTML
		dataMap["AdHref"+n] = ad.ClickHref
	}
	fmt.Printf("data for template: %v", dataMap)

	// Use template from publisher config
	templatePath := "storage/html/" + pubConfig.SerpTemplate
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Fatalf("template parse error: %v", err)
	}
	// execute template
	if err := t.Execute(w, dataMap); err != nil {
		log.Printf("template execute error: %v", err)
	}
}
