package handlers

import "net/http"

func HandleFirstCallJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	http.ServeFile(w, r, "storage/js/firstcall.js")
}
