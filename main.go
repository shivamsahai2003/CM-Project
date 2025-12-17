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
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	if err := db.Init(cfg.DBDsn); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize services
	keywordService := services.NewKeywordService(cfg.APIBaseURL)
	yahooService := services.NewYahooService()
	clickService := services.NewClickService()

	// Initialize handlers
	renderHandler := handlers.NewRenderHandler(keywordService)
	serpHandler := handlers.NewSerpHandler(yahooService)
	adClickHandler := handlers.NewAdClickHandler(clickService)
	keywordHandler := handlers.KeywordsPageHandler{
		KeywordService: keywordService,
	}

	// Register routes
	http.HandleFunc("/firstcall.js", handlers.HandleFirstCallJS)
	http.HandleFunc("/render.js", renderHandler.Handle)
	http.HandleFunc("/serp", serpHandler.Handle)
	http.HandleFunc("/ad-click", adClickHandler.Handle)
	http.HandleFunc("/keywords", keywordHandler.Handle)

	// Start server
	log.Printf("Serving on http://localhost%s ...", cfg.ServerAddr)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr, nil))
}
