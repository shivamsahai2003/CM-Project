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

// AdClickHandler handles ad click tracking and redirects
type AdClickHandler struct {
	clickService *services.ClickService
}

// NewAdClickHandler creates a new ad click handler
func NewAdClickHandler(clickService *services.ClickService) *AdClickHandler {
	return &AdClickHandler{
		clickService: clickService,
	}
}

// Handle processes ad click requests
func (h *AdClickHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()
	if utils.IsBotUA(ua) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "<h2>403 – Bot traffic blocked</h2>")
		return
	}

	qv := r.URL.Query()
	targetRaw := qv.Get("u")
	if targetRaw == "" {
		http.Error(w, "missing u (target) parameter", http.StatusBadRequest)
		return
	}

	target, err := utils.SafeTargetURL(targetRaw)
	if err != nil {
		http.Error(w, "invalid target", http.StatusBadRequest)
		return
	}

	slot := qv.Get("slot")
	kid := utils.AtoiOrZero(qv.Get("kid"))
	q := qv.Get("q")
	adHost := qv.Get("adhost")
	pubID := utils.AtoiOrZero(qv.Get("lid"))
	clientID := utils.GetClientIP(r)

	// In-memory counter
	key := models.ClickStatKey{Slot: slot, KID: strconv.Itoa(kid), Q: q, AdHost: adHost}
	count := h.clickService.IncrementClick(key)
	log.Printf("AD-CLICK: ip=%s ua=%q slot=%q kid=%d q=%q adhost=%q target=%q count=%d", clientID, ua, slot, kid, q, adHost, target, count)

	// DB log
	if pubID > 0 {
		adDetails := fmt.Sprintf(`{"adhost":%q,"target":%q}`, adHost, target)
		_, err := db.GetDB().Exec(
			"INSERT INTO adclick_click (keyword_id, `time`, `user id`, keyword_title, Ad_details, User_agent, publisher_id) VALUES (?, NOW(), ?, ?, ?, ?, ?)",
			kid, clientID, q, adDetails, ua, pubID,
		)
		if err != nil {
			log.Printf("insert adclick_click error: %v", err)
		}
	} else {
		log.Printf("ad click with missing publisher_id (lid)")
	}

	http.Redirect(w, r, target, http.StatusFound)
}
