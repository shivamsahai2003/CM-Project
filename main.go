package main

import (
	"adserving/templates"
	//"adserving/templates"
	//"fmt"
	"log"
	"net/http"
	"regexp"

	"adserving/config"
	"adserving/db"
	"adserving/handlers"
	"adserving/services"
)

func CountAdPlaceHolders(templateStr string) int {
	re := regexp.MustCompile(`\{\{\.AdDesc\d+\}\}`)
	matches := re.FindAllString(templateStr, -1)
	return len(matches)
}

func main() {
	// Load configuration
	cfg := config.Load()
	CountOfAdsFromTemplate := CountAdPlaceHolders(templates.SerpTemplate1)

	// Initialize database connection
	if err := db.Init(cfg.DBDsn); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set DB connection for publisher config
	config.SetDB(db.GetDB())

	// Initialize services
	keywordService := services.NewKeywordService(cfg.APIBaseURL)
	yahooService := services.NewYahooService()
	clickService := services.NewClickService()

	// Initialize handlers
	keywordRenderHandler := handlers.NewRenderHandler(keywordService) // todo rename
	serpHandler := handlers.NewSerpHandler(yahooService)
	adClickHandler := handlers.NewAdClickHandler(clickService)
	//keywordHandler := handlers.KeywordsPageHandler{
	//	KeywordService: keywordService,
	//}

	// Register routes
	http.HandleFunc("/firstcall.js", handlers.HandleFirstCallJS)
	http.HandleFunc("/keyword_render", keywordRenderHandler.Handle)
	http.HandleFunc("/serp", serpHandler.Handle)
	http.HandleFunc("/ad-click", adClickHandler.Handle)
	//http.HandleFunc("/keywords", keywordHandler.Handle) // under-testing

	// Start server
	log.Printf("Serving on http://localhost%s ...", cfg.ServerAddr)
	log.Printf("No of ads showing from template: %d", CountOfAdsFromTemplate)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr, nil))
}
