package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"adserving/db"
	"adserving/models"
	"adserving/services"
	"adserving/utils"
)

type AdClickHandler struct {
	clickService *services.ClickService
}

func NewAdClickHandler(clickService *services.ClickService) *AdClickHandler {
	return &AdClickHandler{clickService: clickService}
}

func (h *AdClickHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userAgent := r.UserAgent()
	isBot := utils.IsBotUA(userAgent)

	q := r.URL.Query()
	targetRaw := q.Get("u")
	if targetRaw == "" {
		http.Error(w, "missing target", http.StatusBadRequest)
		return
	}

	target, err := utils.SafeTargetURL(targetRaw)
	if err != nil {
		http.Error(w, "invalid target", http.StatusBadRequest)
		return
	}

	slot := q.Get("slot")
	keywordID := utils.AtoiOrZero(q.Get("kid"))
	query := q.Get("q")
	adHost := q.Get("adhost")
	adTitle := q.Get("adtitle")
	countryCode := q.Get("cc")
	publisherID := utils.AtoiOrZero(q.Get("pid"))
	clientIP := utils.GetClientIP(r)

	key := models.ClickStatKey{Slot: slot, KeywordID: strconv.Itoa(keywordID), Query: query, AdHost: adHost}
	h.clickService.IncrementClick(key)

	if publisherID > 0 {
		_, err := db.GetDB().Exec(
			`INSERT INTO ad_click (publisher_id, keyword_id, keyword_title, ad_title, ad_host, ad_target_url, slot, client_ip, user_agent, country_code) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			publisherID, keywordID, query, adTitle, adHost, target, slot, clientIP, userAgent, countryCode,
		)
		if err != nil {
			log.Printf("ad_click insert error: %v", err)
		}
	}

	if isBot {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Click logged")
		return
	}

	http.Redirect(w, r, target, http.StatusFound)
}
