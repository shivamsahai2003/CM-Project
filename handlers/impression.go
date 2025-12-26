package handlers

import (
	"log"
	"net/http"
	"strings"

	"adserving/db"
	"adserving/utils"
)

func HandleKeywordImpression(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	q := r.URL.Query()
	publisherID := utils.AtoiOrZero(q.Get("pid"))
	slot := q.Get("slot")
	countryCode := q.Get("cc")
	keywords := q.Get("keywords")
	keywordIDs := q.Get("keyword_ids")

	clientIP := utils.GetClientIP(r)
	userAgent := r.UserAgent()

	keywordList := strings.Split(keywords, ",")
	keywordIDList := strings.Split(keywordIDs, ",")

	for i, kw := range keywordList {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}

		var keywordID any = nil
		if i < len(keywordIDList) {
			if id := utils.AtoiOrZero(strings.TrimSpace(keywordIDList[i])); id != 0 {
				keywordID = id
			}
		}

		_, err := db.GetDB().Exec(
			`INSERT INTO keyword_impression (publisher_id, keyword_id, keyword_title, slot, client_ip, user_agent, country_code) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			publisherID, keywordID, kw, slot, clientIP, userAgent, countryCode,
		)
		if err != nil {
			log.Printf("keyword_impression insert error: %v", err)
		}
	}

	// 1x1 transparent GIF
	gif := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b}
	w.Write(gif)
}
