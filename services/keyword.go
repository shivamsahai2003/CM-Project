package services

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"adserving/models"
)

const defaultAPIBase = "http://g-usw1b-kwd-api-realapi.srv.media.net/kbb/keyword_api.php"

var DefaultKeywords = []string{
	"Best deals online",
	"Top rated products",
	"Compare prices now",
}

var DefaultKeywordIDs = []int64{0, 0, 0}

type KeywordService struct {
	apiBaseURL string
	client     *http.Client
}

func NewKeywordService(apiBaseURL string) *KeywordService {
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBase
	}
	return &KeywordService{
		apiBaseURL: apiBaseURL,
		client:     &http.Client{Timeout: 8 * time.Second},
	}
}

func (s *KeywordService) FetchKeywords(params models.RenderParams) ([]string, []int64, error) {
	q := url.Values{}

	// maxno/actno from template slot count
	maxno := "5"
	if params.MaxNumber > 0 {
		maxno = strconv.Itoa(params.MaxNumber)
	}
	q.Set("maxno", maxno)
	q.Set("actno", maxno)
	q.Set("json", "1")
	q.Set("type", "1")
	q.Set("https", "1")

	// Hardcoded partner/publisher params
	q.Set("csid", "8CUJM46V5")
	q.Set("pid", "8POJDA6W3")
	q.Set("partnerid", "7PRFT79UO")
	q.Set("fpid", "800015395")
	q.Set("crid", "849176236")

	// Feature flags
	q.Set("combineExpired", "1")
	q.Set("fm_skc", "1")
	q.Set("lmsc", "1")
	q.Set("hs", "3")
	q.Set("kf", "0")
	q.Set("kwrd", "0")
	q.Set("py", "1")
	q.Set("pt", "60")
	q.Set("uftr", "0")
	q.Set("ugd", "4")
	q.Set("ykf", "1")
	q.Set("stag_tq_block", "1")
	q.Set("calling_source", "cm")

	// Tags
	q.Set("pstag", "skenzo_test")
	q.Set("stags", "skenzo_test")
	q.Set("mtags", "{perform,BT1_sp},{sem,app,dmsedo,mva,stm,conndigi,pdeal,audext,conn,ginsu}")

	// Dynamic params with defaults
	if params.CountryCode != "" {
		q.Set("cc", params.CountryCode)
	} else {
		q.Set("cc", "US")
	}

	if params.LayoutID != "" {
		q.Set("lid", params.LayoutID)
	} else {
		q.Set("lid", "224")
	}

	if params.TemplateSize != "" {
		q.Set("tsize", params.TemplateSize)
	} else {
		q.Set("tsize", "300x250")
	}

	if params.Domain != "" {
		q.Set("d", params.Domain)
		// Extract TLD (last part after dot, e.g., "com" from "forbes.com")
		parts := strings.Split(params.Domain, ".")
		if len(parts) >= 2 {
			q.Set("dtld", parts[len(parts)-1])
		}
	}

	if params.PageTitle != "" {
		q.Set("ptitle", params.PageTitle)
	}

	if params.ReferrerURL != "" {
		q.Set("rurl", params.ReferrerURL)
	}

	if params.KeywordRef != "" {
		q.Set("kwrf", params.KeywordRef)
	}

	apiURL := s.apiBaseURL + "?" + q.Encode()
	log.Printf("Keyword API URL: %s", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("keyword API request error: %v, using defaults", err)
		return DefaultKeywords, DefaultKeywordIDs, nil
	}
	req.Header.Set("User-Agent", "KeywordService/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("keyword API fetch error: %v, using defaults", err)
		return DefaultKeywords, DefaultKeywordIDs, nil
	}
	defer resp.Body.Close()

	var reader io.ReadCloser = resp.Body
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		if gr, err := gzip.NewReader(resp.Body); err == nil {
			defer gr.Close()
			reader = gr
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("keyword API read error: %v, using defaults", err)
		return DefaultKeywords, DefaultKeywordIDs, nil
	}

	keywords, ids, err := ExtractKeywords(body)
	if err != nil || len(keywords) == 0 {
		log.Printf("keyword API parse error or empty: %v, using defaults", err)
		return DefaultKeywords, DefaultKeywordIDs, nil
	}

	return keywords, ids, nil
}

func ExtractKeywords(body []byte) ([]string, []int64, error) {
	var resp models.KeywordResponse
	if err := json.Unmarshal(body, &resp); err == nil && len(resp.Keywords) > 0 {
		var keywords []string
		var ids []int64
		for _, item := range resp.Keywords {
			text := strings.TrimSpace(item.Title)
			if text == "" {
				continue
			}
			keywords = append(keywords, text)
			var id int64
			if item.ID != "" {
				id, _ = item.ID.Int64()
			}
			ids = append(ids, id)
		}
		return keywords, ids, nil
	}

	var root any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return nil, nil, fmt.Errorf("json decode: %w", err)
	}

	var keywords []string
	var ids []int64
	seen := map[string]struct{}{}

	toInt64 := func(v any) int64 {
		switch t := v.(type) {
		case json.Number:
			i, _ := t.Int64()
			return i
		case float64:
			return int64(t)
		case int:
			return int64(t)
		case int64:
			return t
		case string:
			i, _ := strconv.ParseInt(t, 10, 64)
			return i
		}
		return 0
	}

	toString := func(v any) string {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}

	var walk func(any)
	walk = func(v any) {
		switch t := v.(type) {
		case []any:
			for _, item := range t {
				walk(item)
			}
		case map[string]any:
			if kv, ok := t["k"]; ok {
				if arr, ok := kv.([]any); ok {
					for _, obj := range arr {
						if m, ok := obj.(map[string]any); ok {
							text := strings.TrimSpace(toString(m["t"]))
							if text != "" {
								if _, exists := seen[text]; !exists {
									seen[text] = struct{}{}
									keywords = append(keywords, text)
									ids = append(ids, toInt64(m["i"]))
								}
							}
						}
					}
				}
			}
			for _, nv := range t {
				walk(nv)
			}
		}
	}
	walk(root)

	return keywords, ids, nil
}
