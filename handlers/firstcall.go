package handlers

import (
	"fmt"
	"net/http"
)

// HandleFirstCallJS serves the first call JavaScript
func HandleFirstCallJS(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	//fmt.Fprint(w, templates.FirstCallJS) // todo fix this
	fmt.Println("entered 1st call")

	// Disable caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeFile(w, r, "storage/js/firstcall.js")
}
