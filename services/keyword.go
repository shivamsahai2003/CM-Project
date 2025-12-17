package services

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"adserving/models"
)

const defaultAPIBase = "http://g-usw1b-kwd-api-realapi.srv.media.net/kbb/keyword_api.php"

// KeywordService handles keyword-related operations
type KeywordService struct {
	apiBaseURL string
	client     *http.Client
}

// NewKeywordService creates a new keyword service
func NewKeywordService(apiBaseURL string) *KeywordService {
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBase
	}
	return &KeywordService{
		apiBaseURL: apiBaseURL,
		client:     &http.Client{Timeout: 8 * time.Second},
	}
}

// FetchKeywords fetches keywords from the API
func (s *KeywordService) FetchKeywords(params models.RenderParams) ([]string, []int64, error) {
	vals := url.Values{}

	if params.Actno != "" {
		vals.Set("actno", params.Actno)
	} else {
		vals.Set("actno", "5")
	}
	if params.Maxno != "" {
		vals.Set("maxno", params.Maxno)
	} else {
		vals.Set("maxno", "5")
	}
	if params.CC != "" {
		vals.Set("cc", params.CC)
	} else {
		vals.Set("cc", "US")
	}
	if params.LID != "" {
		vals.Set("lid", params.LID)
	} else {
		vals.Set("lid", "224")
	}
	if params.D != "" {
		vals.Set("d", params.D)
	}
	if params.RURL != "" {
		vals.Set("rurl", params.RURL)
	}
	if params.PTitle != "" {
		vals.Set("ptitle", params.PTitle)
	}
	if params.TSize != "" {
		vals.Set("tsize", params.TSize)
	} else {
		vals.Set("tsize", "300x250")
	}
	if params.KwRf != "" {
		vals.Set("kwrf", params.KwRf)
	}
	vals.Set("json", "1")

	apiURL := s.apiBaseURL + "?" + vals.Encode()
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "KeywordService/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("keyword API fetch error: %w", err)
	}
	defer resp.Body.Close()

	var reader io.ReadCloser = resp.Body
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		gzr, gzErr := gzip.NewReader(resp.Body)
		if gzErr == nil {
			defer gzr.Close()
			reader = gzr
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("keyword API read error: %w", err)
	}

	return ExtractKeywords(body)
}

// ExtractKeywords parses keywords from the API response body
func ExtractKeywords(body []byte) ([]string, []int64, error) {
	var resp models.KeywordResponse
	if err := json.Unmarshal(body, &resp); err == nil && len(resp.K) > 0 {
		kws := make([]string, 0, len(resp.K))
		ids := make([]int64, 0, len(resp.K))
		for _, it := range resp.K {
			txt := strings.TrimSpace(it.T)
			if txt == "" {
				txt = strings.TrimSpace(it.DT)
			}
			if txt == "" {
				continue
			}
			kws = append(kws, txt)
			var id int64
			if it.I != "" {
				if v, err := it.I.Int64(); err == nil {
					id = v
				}
			}
			ids = append(ids, id)
		}
		return kws, ids, nil
	}

	// Fallback tolerant walker
	var root any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return nil, nil, fmt.Errorf("json decode: %w", err)
	}

	var kws []string
	var ids []int64
	seen := map[string]struct{}{}

	toInt64 := func(v any) int64 {
		switch n := v.(type) {
		case json.Number:
			if i, err := n.Int64(); err == nil {
				return i
			}
		case float64:
			return int64(n)
		case float32:
			return int64(n)
		case int:
			return int64(n)
		case int64:
			return n
		case string:
			if i, err := strconv.ParseInt(n, 10, 64); err == nil {
				return i
			}
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
			for _, it := range t {
				walk(it)
			}
		case map[string]any:
			if kv, ok := t["k"]; ok {
				if arr, ok := kv.([]any); ok {
					for _, ko := range arr {
						if kwObj, ok := ko.(map[string]any); ok {
							txt := strings.TrimSpace(toString(kwObj["t"]))
							if txt == "" {
								txt = strings.TrimSpace(toString(kwObj["dt"]))
							}
							if txt != "" {
								if _, exists := seen[txt]; !exists {
									seen[txt] = struct{}{}
									kws = append(kws, txt)
									ids = append(ids, toInt64(kwObj["i"]))
								}
							}
						}
					}
				}
			}
			for _, v2 := range t {
				walk(v2)
			}
		}
	}
	walk(root)

	return kws, ids, nil
}
