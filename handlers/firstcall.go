package handlers

import (
	"fmt"
	"net/http"

	"adserving/templates"
)

// HandleFirstCallJS serves the first call JavaScript
func HandleFirstCallJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	fmt.Fprint(w, templates.FirstCallJS)
}
