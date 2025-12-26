package handlers

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"adserving/config"
	"adserving/models"
	"adserving/services"
	"adserving/utils"
)

const dummyKeywordTemplate = "storage/html/KeywordTemplateDummy.html"

type RenderHandler struct {
	keywordService *services.KeywordService
}

func NewRenderHandler(keywordService *services.KeywordService) *RenderHandler {
	return &RenderHandler{keywordService: keywordService}
}

func (h *RenderHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userAgent := r.UserAgent()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	q := r.URL.Query()
	params := models.RenderParams{
		Slot:         q.Get("slot"),
		CountryCode:  q.Get("cc"),
		TemplateSize: q.Get("tsize"),
		PublisherID:  q.Get("pid"),
		Domain:       q.Get("d"),
		LayoutID:     q.Get("lid"),
		PageTitle:    q.Get("ptitle"),
		ReferrerURL:  q.Get("rurl"),
		KeywordRef:   q.Get("kwrf"),
	}

	if params.Slot == "" {
		renderErrorHTML(w, "No slot provided")
		return
	}

	publisherID := utils.AtoiOrZero(params.PublisherID)
	rule := config.GetRuleByPublisherIDAndUserAgent(publisherID, userAgent)

	if rule.Action.Block {
		w.WriteHeader(http.StatusForbidden)
		renderErrorHTML(w, "Traffic blocked")
		return
	}

	// Try to get template, fallback to dummy
	keywordTemplatePath := "storage/html/" + rule.Action.KeywordTemplateID
	maxKeywords := utils.CountKeywordSlots(keywordTemplatePath)
	if maxKeywords == 0 {
		keywordTemplatePath = dummyKeywordTemplate
		maxKeywords = 3
	}

	// Set maxno from template slot count
	params.MaxNumber = maxKeywords

	// FetchKeywords returns defaults on error
	keywords, keywordIDs, _ := h.keywordService.FetchKeywords(params)

	if len(keywords) > maxKeywords {
		keywords = keywords[:maxKeywords]
		if len(keywordIDs) > maxKeywords {
			keywordIDs = keywordIDs[:maxKeywords]
		}
	}

	baseURL := utils.GetScheme(r) + "://" + r.Host

	linkTarget := "_parent"
	if rule.Action.OpenInNewTab {
		linkTarget = "_blank"
	}

	dataMap := map[string]interface{}{"LinkTarget": linkTarget}

	for i, kw := range keywords {
		qs := url.Values{}
		qs.Set("q", kw)
		qs.Set("slot", params.Slot)
		qs.Set("cc", params.CountryCode)
		qs.Set("pid", params.PublisherID)
		if i < len(keywordIDs) && keywordIDs[i] != 0 {
			qs.Set("kid", strconv.FormatInt(keywordIDs[i], 10))
		}

		idx := strconv.Itoa(i + 1)
		dataMap["KwTitle"+idx] = html.EscapeString(kw)
		dataMap["KwHref"+idx] = baseURL + "/serp?" + qs.Encode()
	}

	tmpl, err := template.ParseFiles(keywordTemplatePath)
	if err != nil {
		log.Printf("template parse error: %v, trying dummy", err)
		tmpl, err = template.ParseFiles(dummyKeywordTemplate)
		if err != nil {
			renderErrorHTML(w, "Template error")
			return
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, dataMap); err != nil {
		log.Printf("template execute error: %v", err)
		renderErrorHTML(w, "Template error")
		return
	}

	widthPx, heightPx := utils.ParseSize(params.TemplateSize)
	if widthPx <= 0 {
		widthPx = 300
	}
	if heightPx <= 0 {
		heightPx = 250
	}

	var kidStrs []string
	for _, id := range keywordIDs {
		kidStrs = append(kidStrs, strconv.FormatInt(id, 10))
	}

	impParams := url.Values{}
	impParams.Set("pid", params.PublisherID)
	impParams.Set("slot", params.Slot)
	impParams.Set("cc", params.CountryCode)
	impParams.Set("keywords", strings.Join(keywords, ","))
	impParams.Set("keyword_ids", strings.Join(kidStrs, ","))
	impURL := baseURL + "/keyword_impression?" + impParams.Encode()

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{width:%dpx;height:%dpx;border:1px solid #e2e8f0;border-radius:8px;background:#fff;overflow:auto;padding:8px;font-family:Arial,sans-serif}
a{color:#1a73e8;text-decoration:none}
a:hover{text-decoration:underline}
</style>
</head>
<body>
%s
<script>if(window.parent!==window){window.parent.postMessage({type:'impression',url:'%s'},'*');}</script>
</body>
</html>`, widthPx, heightPx, buf.String(), impURL)
}

func renderErrorHTML(w http.ResponseWriter, msg string) {
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"></head><body style="margin:0;padding:8px;font:14px Arial;color:#555">%s</body></html>`, html.EscapeString(msg))
}
