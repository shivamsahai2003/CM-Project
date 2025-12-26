package main

import (
	"log"
	"net/http"

	"adserving/config"
	"adserving/db"
	"adserving/handlers"
	"adserving/services"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg.DBDsn); err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	defer db.Close()

	config.SetRulesDB(db.GetDB())

	keywordService := services.NewKeywordService(cfg.APIBaseURL)
	yahooService := services.NewYahooService()
	clickService := services.NewClickService()

	renderHandler := handlers.NewRenderHandler(keywordService)
	serpHandler := handlers.NewSerpHandler(yahooService)
	adClickHandler := handlers.NewAdClickHandler(clickService)

	http.HandleFunc("/firstcall.js", handlers.HandleFirstCallJS)
	http.HandleFunc("/keyword_render", renderHandler.Handle)
	http.HandleFunc("/keyword_impression", handlers.HandleKeywordImpression)
	http.HandleFunc("/serp", serpHandler.Handle)
	http.HandleFunc("/ad-click", adClickHandler.Handle)

	log.Printf("Server starting on %s", cfg.ServerAddr)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr, nil))
}
