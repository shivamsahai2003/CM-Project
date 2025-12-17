// handlers/keywords.go
package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"adserving/models"
	"adserving/services"
	"adserving/templates"
)

type KeywordItem struct {
	ID   int64
	Name string
}

type KeywordGroup struct {
	Label    string
	Keywords []KeywordItem
}

type KeywordsPageData struct {
	Title        string
	PubKey       string
	TotalFetched int
	TotalShown   int
	Groups       []KeywordGroup

	// SERP-like metadata
	Slot   string
	CC     string
	D      string
	LID    int
	PID    int
	TSize  string
	KwRf   string
	PTitle string
	RURL   string
	KID    int
}

type KeywordsPageHandler struct {
	KeywordService *services.KeywordService
}

// Per-publisher keyword counts for the 4 divs.
var pubAlloc = map[string][]int{
	"blue":    {2, 3, 1, 2},
	"red":     {5, 4, 3, 2},
	"default": {2, 2, 1, 1},
}

// Optional: per-publisher meta (PID, CC). Adjust to your real config or pull from your config package.
var pubMeta = map[string]struct {
	PID int
	CC  string
}{
	"blue":    {PID: 100, CC: "US"},
	"red":     {PID: 200, CC: "US"},
	"default": {PID: 0, CC: "US"},
}

func sumAlloc(pubKey string) int {
	a, ok := pubAlloc[pubKey]
	if !ok {
		a = pubAlloc["default"]
	}
	s := 0
	for i := 0; i < 4 && i < len(a); i++ {
		s += a[i]
	}
	return s
}

func buildGroups(pubKey string, names []string, ids []int64) ([]KeywordGroup, int) {
	a, ok := pubAlloc[pubKey]
	if !ok {
		a = pubAlloc["default"]
	}
	labels := []string{"Div 1", "Div 2", "Div 3", "Div 4"}

	groups := make([]KeywordGroup, 0, 4)
	idx := 0
	shown := 0
	for i := 0; i < 4; i++ {
		need := 0
		if i < len(a) {
			need = a[i]
		}
		end := idx + need
		if end > len(names) {
			end = len(names)
		}
		g := KeywordGroup{Label: labels[i]}
		for j := idx; j < end; j++ {
			item := KeywordItem{Name: names[j]}
			if j < len(ids) {
				item.ID = ids[j]
			}
			g.Keywords = append(g.Keywords, item)
		}
		if idx < end {
			shown += end - idx
			idx = end
		}
		groups = append(groups, g)
	}
	return groups, shown
}

func requestURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xf := r.Header.Get("X-Forwarded-Proto"); xf != "" {
		scheme = xf
	}
	return scheme + "://" + r.Host + r.URL.RequestURI()
}

var keywordsTmpl = template.Must(template.New("keywords").Parse(templates.KeywordTemplate))

// FIX: use the correct receiver type and parameter names (no underscores)
func (h *KeywordsPageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Identify publisher from ?pub= or host
	pubKey := r.URL.Query().Get("pub")
	if pubKey == "" {
		host := strings.TrimPrefix(strings.Split(r.Host, ":")[0], "www.")
		switch host {
		case "blue.localhost":
			pubKey = "blue"
		case "red.localhost":
			pubKey = "red"
		default:
			pubKey = "default"
		}
	}

	// Publisher meta
	meta, ok := pubMeta[pubKey]
	if !ok {
		meta = pubMeta["default"]
	}

	// Decide how many keywords to fetch/show
	wanted := sumAlloc(pubKey)

	// Build params for KeywordService and fetch from API
	params := models.RenderParams{
		Pub:    pubKey,
		Maxno:  strconv.Itoa(wanted), // also used as "limit" in the service
		Actno:  "5",
		CC:     meta.CC,
		LID:    strconv.Itoa(224),
		TSize:  "300x250",
		KwRf:   r.Referer(),
		RURL:   requestURL(r),
		PTitle: "Publisher Keywords",
		D:      strings.TrimPrefix(strings.Split(r.Host, ":")[0], "www."),
	}

	names, ids, err := h.KeywordService.FetchKeywords(params)
	if err != nil {
		http.Error(w, "failed to fetch keywords: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Trim locally if API returns more
	if len(names) > wanted {
		names = names[:wanted]
		if len(ids) >= wanted {
			ids = ids[:wanted]
		}
	}

	groups, shown := buildGroups(pubKey, names, ids)

	data := KeywordsPageData{
		Title:        "Publisher Keywords",
		PubKey:       pubKey,
		TotalFetched: len(names),
		TotalShown:   shown,
		Groups:       groups,

		// SERP-like metadata
		Slot:   "keywords",
		CC:     meta.CC,
		D:      params.D,
		LID:    224,
		PID:    meta.PID,
		TSize:  "300x250",
		KwRf:   params.KwRf,
		PTitle: params.PTitle,
		RURL:   params.RURL,
		KID:    0,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := keywordsTmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
