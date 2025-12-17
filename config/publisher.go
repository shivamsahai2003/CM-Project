package config

// PublisherConfig holds configuration for a publisher
type PublisherConfig struct {
	PID    int    // Publisher ID
	LID    int    // Layout ID
	Actno  int    // Active number of ads
	Maxno  int    // Maximum number of ads
	CC     string // Country code
	TSize  string // Tile size
	MaxAds int    // Max ads on SERP
}

// DefaultPublisherConfig is used when publisher is not found
var DefaultPublisherConfig = PublisherConfig{
	PID:    0,
	LID:    224,
	Actno:  5,
	Maxno:  5,
	CC:     "US",
	TSize:  "300x250",
	MaxAds: 3,
}

// PubKeyConfig maps publisher key (e.g., "blue", "red") to configuration
var PubKeyConfig = map[string]PublisherConfig{
	"blue": {
		PID:    100,
		LID:    224,
		Actno:  5,
		Maxno:  5,
		CC:     "US",
		TSize:  "300x250",
		MaxAds: 2, // Blue Publisher - 2 ads on SERP
	},
	"red": {
		PID:    200,
		LID:    224,
		Actno:  5,
		Maxno:  5,
		CC:     "US",
		TSize:  "300x250",
		MaxAds: 5, // Red Publisher - 5 ads on SERP
	},
}

// HostPublisherConfig maps host/domain to publisher configuration (fallback)
var HostPublisherConfig = map[string]PublisherConfig{
	"blue.localhost": PubKeyConfig["blue"],
	"red.localhost":  PubKeyConfig["red"],
	"localhost": {
		PID:    0,
		LID:    224,
		Actno:  5,
		Maxno:  5,
		CC:     "US",
		TSize:  "300x250",
		MaxAds: 3,
	},
}

// GetPublisherConfigByPubKey returns publisher config for a given pub key (e.g., "blue", "red")
func GetPublisherConfigByPubKey(pubKey string) PublisherConfig {
	if cfg, exists := PubKeyConfig[pubKey]; exists {
		return cfg
	}
	return DefaultPublisherConfig
}

// GetPublisherConfigByHost returns publisher config for a given host (fallback)
func GetPublisherConfigByHost(host string) PublisherConfig {
	if cfg, exists := HostPublisherConfig[host]; exists {
		return cfg
	}
	return DefaultPublisherConfig
}

// GetPublisherConfig returns config by pub key first, then by host as fallback
func GetPublisherConfig(pubKey, host string) PublisherConfig {
	if pubKey != "" && pubKey != "default" {
		if cfg, exists := PubKeyConfig[pubKey]; exists {
			return cfg
		}
	}
	return GetPublisherConfigByHost(host)
}
