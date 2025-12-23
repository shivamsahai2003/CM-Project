package config

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

// PublisherConfig holds configuration for a publisher
type PublisherConfig struct {
	PID             int    // Publisher ID
	LID             int    // Layout ID
	CC              string // Country code
	TSize           string // Tile size
	SerpTemplate    string // SERP template file name
	KeywordTemplate string // Keyword template file name
}

// DefaultPublisherConfig is used when publisher is not found
var DefaultPublisherConfig = PublisherConfig{
	PID:             0,
	LID:             224,
	CC:              "US",
	TSize:           "300x250",
	SerpTemplate:    "SerpTemplate1.html",
	KeywordTemplate: "KeywordTemplate1.html",
}

// Cache for publisher configs (cleared every 5 minutes)
var (
	configCache = make(map[int]PublisherConfig)
	cacheMutex  sync.RWMutex
	dbConn      *sql.DB
	cacheOnce   sync.Once
)

// SetDB sets the database connection for fetching publisher configs
func SetDB(db *sql.DB) {
	dbConn = db

	// Start cache cleanup goroutine (only once)
	cacheOnce.Do(func() {
		go startCacheCleanup()
	})
}

// startCacheCleanup clears the cache every 5 minutes
func startCacheCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		cacheMutex.Lock()
		configCache = make(map[int]PublisherConfig)
		cacheMutex.Unlock()
		log.Println("Publisher config cache cleared (5 min TTL)")
	}
}

// GetPublisherConfigByPID returns publisher config for a given PID from database
// Uses caching with 5 minute TTL
func GetPublisherConfigByPID(pid int) PublisherConfig {
	if pid == 0 {
		return DefaultPublisherConfig
	}

	// Check cache first
	cacheMutex.RLock()
	if cfg, exists := configCache[pid]; exists {
		cacheMutex.RUnlock()
		return cfg
	}
	cacheMutex.RUnlock()

	// Fetch from database
	if dbConn == nil {
		log.Printf("Database connection not set, returning default config for PID %d", pid)
		return DefaultPublisherConfig
	}

	var cfg PublisherConfig
	err := dbConn.QueryRow(`
		SELECT pid, lid, cc, tsize, serp_template, keyword_template 
		FROM pub_config 
		WHERE pid = ?
	`, pid).Scan(
		&cfg.PID,
		&cfg.LID,
		&cfg.CC,
		&cfg.TSize,
		&cfg.SerpTemplate,
		&cfg.KeywordTemplate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No config found for PID %d, returning default", pid)
		} else {
			log.Printf("Error fetching config for PID %d: %v", pid, err)
		}
		return DefaultPublisherConfig
	}

	// Cache the result
	cacheMutex.Lock()
	configCache[pid] = cfg
	cacheMutex.Unlock()

	return cfg
}

// ClearConfigCache manually clears the publisher config cache
func ClearConfigCache() {
	cacheMutex.Lock()
	configCache = make(map[int]PublisherConfig)
	cacheMutex.Unlock()
	log.Println("Publisher config cache manually cleared")
}

// UpsertPublisherConfig inserts or updates a publisher config in the database
func UpsertPublisherConfig(cfg PublisherConfig) error {
	if dbConn == nil {
		log.Printf("Database connection not set")
		return nil
	}

	_, err := dbConn.Exec(`
		INSERT INTO pub_config (pid, lid, cc, tsize, serp_template, keyword_template)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			lid = VALUES(lid),
			cc = VALUES(cc),
			tsize = VALUES(tsize),
			serp_template = VALUES(serp_template),
			keyword_template = VALUES(keyword_template)
	`, cfg.PID, cfg.LID, cfg.CC, cfg.TSize, cfg.SerpTemplate, cfg.KeywordTemplate)

	if err != nil {
		log.Printf("Error upserting config for PID %d: %v", cfg.PID, err)
		return err
	}

	// Clear cache for this PID
	cacheMutex.Lock()
	delete(configCache, cfg.PID)
	cacheMutex.Unlock()

	return nil
}
