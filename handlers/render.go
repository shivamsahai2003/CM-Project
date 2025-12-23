package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	"adserving/utils"
)

// RenderHandler handles render.js requests
type RenderHandler struct {
	keywordService *services.KeywordService
}

// NewRenderHandler creates a new render handler
func NewRenderHandler(keywordService *services.KeywordService) *RenderHandler {
	return &RenderHandler{
		keywordService: keywordService,
	}
}

// Handle processes the render.js request

func (h *RenderHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()
	if utils.IsBotUA(ua) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "// Bot traffic blocked")
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8") // todo check code sequence

	incomingQueryParams := r.URL.Query()

	params := models.RenderParams{
		Slot:   incomingQueryParams.Get("slot"),
		Maxno:  incomingQueryParams.Get("maxno"),
		CC:     incomingQueryParams.Get("cc"),
		LID:    incomingQueryParams.Get("lid"),
		D:      incomingQueryParams.Get("d"),
		RURL:   incomingQueryParams.Get("rurl"),
		PTitle: incomingQueryParams.Get("ptitle"),
		TSize:  incomingQueryParams.Get("tsize"),
		KwRf:   incomingQueryParams.Get("kwrf"),
		PID:    incomingQueryParams.Get("pid"),
	}

	if params.Slot == "" {
		fmt.Fprint(w, `(function(){ /* no slot provided */ })();`)
		return
	}

	// Get publisher config by PID BEFORE fetching keywords
	pid := utils.AtoiOrZero(params.PID)
	pubConfig := config.GetPublisherConfigByPID(pid)

	// Get keyword template and count slots to determine maxno
	keywordTemplatePath := "storage/html/" + pubConfig.KeywordTemplate
	maxKeywords := utils.CountKeywordSlots(keywordTemplatePath)
	if maxKeywords == 0 {
		maxKeywords = 6 // fallback
	}

	// Override maxno based on keyword template slots
	params.Maxno = strconv.Itoa(maxKeywords)
	log.Printf("Render: PID=%d, KeywordTemplate=%s, MaxKeywords=%d", pid, pubConfig.KeywordTemplate, maxKeywords)

	keywords, ids, err := h.keywordService.FetchKeywords(params)
	if err != nil {
		log.Printf("keyword API error: %v", err)
		failJS(w, params.Slot, "Failed to fetch keywords")
		return
	}

	if len(keywords) == 0 {
		// todo some kind of default keyword in case something goes wrong
		fmt.Fprintf(w, `(function(){ var el = document.getElementById(%q); if(!el) return; el.innerHTML = '<div style="font:14px Arial;color:#555;">No keywords available</div>'; })();`, params.Slot)
		return
	}

	fmt.Println("showing keywords: ", keywords)

	// todo
	if params.Maxno != "" {
		if n, err := strconv.Atoi(params.Maxno); err == nil && n > 0 && n < len(keywords) {
			keywords = keywords[:n]
			if len(ids) >= n {
				ids = ids[:n]
			}
		}
	}

	// Ensure publisher row exists, then log impressions
	pubID := utils.AtoiOrZero(params.LID)

	//
	if pubID > 0 && params.D != "" {
		_, _ = db.GetDB().Exec(
			"INSERT INTO publisher (publisher_id, domain) VALUES (?, ?) ON DUPLICATE KEY UPDATE domain=VALUES(domain)",
			pubID, params.D,
		)
	}

	var slotSQL any = nil
	if sInt, err := strconv.Atoi(strings.TrimSpace(params.Slot)); err == nil {
		slotSQL = sInt
	}
	for i, kw := range keywords {
		var kid any = nil
		if i < len(ids) && ids[i] != 0 {
			kid = ids[i]
		}
		_, err := db.GetDB().Exec(
			"INSERT INTO keyword_impression (publisher_id, keyword_no, keywords, slot, user_agent) VALUES (?, ?, ?, ?, ?)",
			pubID, kid, kw, slotSQL, ua,
		)
		if err != nil {
			log.Printf("insert keyword_impression error: %v", err)
		}
	}

	// Render the keyword links
	base := utils.GetScheme(r) + "://" + r.Host    // todo recheck
	wpx, heightpx := utils.ParseSize(params.TSize) // todo recheck
	if wpx <= 0 {
		wpx = 300
	}
	if heightpx <= 0 {
		heightpx = 250
	}

	// Get max ads from SERP template
	serpTemplatePath := "storage/html/" + pubConfig.SerpTemplate
	maxAds := utils.CountAdSlots(serpTemplatePath)
	if maxAds == 0 {
		maxAds = 3 // fallback
	}
	maxAdsStr := strconv.Itoa(maxAds)

	// Limit keywords to the number of slots in the template (already counted above)
	if len(keywords) > maxKeywords {
		keywords = keywords[:maxKeywords]
		if len(ids) > maxKeywords {
			ids = ids[:maxKeywords]
		}
	}

	// Build data map for keyword template
	dataMap := make(map[string]interface{})
	for i, keywordTitle := range keywords {
		qs := url.Values{}
		qs.Set("q", keywordTitle)
		qs.Set("slot", params.Slot)
		if params.CC != "" {
			qs.Set("cc", params.CC)
		}
		if params.D != "" {
			qs.Set("d", params.D)
		}
		if params.RURL != "" {
			qs.Set("rurl", params.RURL)
		}
		if params.PTitle != "" {
			qs.Set("ptitle", params.PTitle)
		}
		if params.LID != "" {
			qs.Set("lid", params.LID)
		}
		if params.TSize != "" {
			qs.Set("tsize", params.TSize)
		}
		if params.KwRf != "" {
			qs.Set("kwrf", params.KwRf)
		}
		qs.Set("pid", params.PID)
		qs.Set("maxads", maxAdsStr)
		if i < len(ids) && ids[i] != 0 {
			qs.Set("kid", strconv.FormatInt(ids[i], 10))
		}
		href := base + "/serp?" + qs.Encode()

		n := strconv.Itoa(i + 1)
		dataMap["KwTitle"+n] = html.EscapeString(keywordTitle)
		dataMap["KwHref"+n] = href
	}

	// Parse and execute keyword template
	t, err := template.ParseFiles(keywordTemplatePath)
	if err != nil {
		log.Printf("Failed to parse keyword template %s: %v", keywordTemplatePath, err)
		failJS(w, params.Slot, "Template error")
		return
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, dataMap); err != nil {
		log.Printf("Failed to execute keyword template: %v", err)
		failJS(w, params.Slot, "Template error")
		return
	}

	// Wrap in a styled container
	htmlContent := fmt.Sprintf(`<div class="kw-box" style="box-sizing:border-box; width:%dpx; height:%dpx; border:1px solid #e2e8f0; border-radius:8px; background:#ffffff; overflow:auto; padding:8px;">%s</div>`,
		wpx, heightpx, buf.String())

	htmlJSON, _ := json.Marshal(htmlContent)
	fmt.Fprintf(w, `(function(){ var el = document.getElementById(%q); if(!el) return; el.innerHTML = %s; })();`, params.Slot, string(htmlJSON))
}

func failJS(w http.ResponseWriter, slot, msg string) {
	msg = html.EscapeString(msg)
	fmt.Fprintf(w, `(function(){ var el = document.getElementById(%q); if(!el) return; el.innerHTML = '<div style="font:14px Arial;color:#b00;">%s</div>'; })();`, slot, msg)
}
