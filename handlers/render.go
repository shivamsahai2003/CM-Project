package handlers

import (
	"encoding/json"
	"fmt"
	"html"
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

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	q := r.URL.Query()

	params := models.RenderParams{
		Slot:   q.Get("slot"),
		Actno:  q.Get("actno"),
		Maxno:  q.Get("maxno"),
		CC:     q.Get("cc"),
		LID:    q.Get("lid"),
		D:      q.Get("d"),
		RURL:   q.Get("rurl"),
		PTitle: q.Get("ptitle"),
		TSize:  q.Get("tsize"),
		KwRf:   q.Get("kwrf"),
		PID:    q.Get("pid"),
		Pub:    q.Get("pub"),
	}

	if params.Slot == "" {
		fmt.Fprint(w, `(function(){ /* no slot provided */ })();`)
		return
	}

	keywords, ids, err := h.keywordService.FetchKeywords(params)
	if err != nil {
		log.Printf("keyword API error: %v", err)
		failJS(w, params.Slot, "Failed to fetch keywords")
		return
	}

	if len(keywords) == 0 {
		fmt.Fprintf(w, `(function(){ var el = document.getElementById(%q); if(!el) return; el.innerHTML = '<div style="font:14px Arial;color:#555;">No keywords available</div>'; })();`, params.Slot)
		return
	}

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
	base := utils.GetScheme(r) + "://" + r.Host
	wpx, hpx := utils.ParseSize(params.TSize)
	if wpx <= 0 {
		wpx = 300
	}
	if hpx <= 0 {
		hpx = 250
	}

	// Get publisher config by pub key (or fallback to domain) to determine max ads for SERP
	pubConfig := config.GetPublisherConfig(params.Pub, params.D)
	maxAdsStr := strconv.Itoa(pubConfig.MaxAds)
	log.Printf("Render: Pub=%s, Domain=%s, MaxAds=%d", params.Pub, params.D, pubConfig.MaxAds)

	var sb strings.Builder
	sb.WriteString(`<div class="kw-box" style="box-sizing:border-box; width:` + strconv.Itoa(wpx) + `px; height:` + strconv.Itoa(hpx) + `px; border:1px solid #e2e8f0; border-radius:8px; background:#ffffff; overflow:auto; padding:8px; display:flex; flex-direction:column; gap:6px;">`)

	for i, t := range keywords {
		qs := url.Values{}
		qs.Set("q", t)
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
		qs.Set("pid", params.PID)   // Always pass PID to SERP
		qs.Set("maxads", maxAdsStr) // Pass max ads from publisher config
		if i < len(ids) && ids[i] != 0 {
			qs.Set("kid", strconv.FormatInt(ids[i], 10))
		}
		href := base + "/serp?" + qs.Encode()
		sb.WriteString(`<a href="` + href + `" style="display:block; padding:8px 10px; border:1px solid #cbd5e1; border-radius:6px; background:#f8fafc; color:#0b57d0; text-decoration:none; font:14px/1.4 Arial, sans-serif;">` + html.EscapeString(t) + `</a>`)
	}
	sb.WriteString(`</div>`)

	htmlJSON, _ := json.Marshal(sb.String())
	fmt.Fprintf(w, `(function(){ var el = document.getElementById(%q); if(!el) return; el.innerHTML = %s; })();`, params.Slot, string(htmlJSON))
}

func failJS(w http.ResponseWriter, slot, msg string) {
	msg = html.EscapeString(msg)
	fmt.Fprintf(w, `(function(){ var el = document.getElementById(%q); if(!el) return; el.innerHTML = '<div style="font:14px Arial;color:#b00;">%s</div>'; })();`, slot, msg)
}
