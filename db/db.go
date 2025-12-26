package db

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// DB holds the database connection pool
var DB *sql.DB

// Init initializes the database connection, creating the database and tables if needed
func Init(dsn string) error {
	dbName, baseDSN := parseDSN(dsn)

	// First connect without database to create it if needed
	if dbName != "" && baseDSN != "" {
		if err := ensureDatabase(baseDSN, dbName); err != nil {
			return fmt.Errorf("failed to ensure database: %w", err)
		}
	}

	// Now connect to the actual database
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Connected to MySQL")

	// Create tables if they don't exist
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Seed default rules
	if err := seedRules(); err != nil {
		return fmt.Errorf("failed to seed rules: %w", err)
	}

	return nil
}

// parseDSN extracts database name and returns DSN without database
// DSN format: user:pass@tcp(host:port)/dbname?params
func parseDSN(dsn string) (dbName, baseDSN string) {
	re := regexp.MustCompile(`^(.+)/([^?]+)(\?.*)?$`)
	matches := re.FindStringSubmatch(dsn)
	if len(matches) >= 3 {
		dbName = matches[2]
		baseDSN = matches[1] + "/" + matches[3] // without db name but with params
		if strings.HasSuffix(baseDSN, "/") {
			baseDSN = baseDSN[:len(baseDSN)-1] + "/?" + strings.TrimPrefix(matches[3], "?")
		}
	}
	return
}

// ensureDatabase creates the database if it doesn't exist
func ensureDatabase(baseDSN, dbName string) error {
	conn, err := sql.Open("mysql", baseDSN)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		return err
	}

	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		return err
	}

	log.Printf("Ensured database '%s' exists", dbName)
	return nil
}

// createTables creates required tables if they don't exist
func createTables() error {
	tables := []string{
		// Publisher table - stores publisher info (100=blue, 200=red)
		`CREATE TABLE IF NOT EXISTS publisher (
			publisher_id INT PRIMARY KEY,
			domain VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		// Rules table - stores targeting rules
		`CREATE TABLE IF NOT EXISTS rules (
			id INT AUTO_INCREMENT PRIMARY KEY,
			rule_name VARCHAR(255) NOT NULL,
			action JSON NOT NULL,
			publisher_id INT NOT NULL,
			user_agent VARCHAR(255) DEFAULT NULL,
			country_code VARCHAR(10) DEFAULT 'US',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_publisher_rule (publisher_id, rule_name)
		)`,
		// Keyword impression - records when keywords are shown on publisher page
		`CREATE TABLE IF NOT EXISTS keyword_impression (
			id INT AUTO_INCREMENT PRIMARY KEY,
			publisher_id INT NOT NULL,
			keyword_id INT,
			keyword_title VARCHAR(500),
			slot VARCHAR(100),
			client_ip VARCHAR(100),
			user_agent TEXT,
			country_code VARCHAR(10),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_publisher_id (publisher_id),
			INDEX idx_created_at (created_at)
		)`,
		// Keyword click - records when a keyword is clicked (redirects to SERP)
		`CREATE TABLE IF NOT EXISTS keyword_click (
			id INT AUTO_INCREMENT PRIMARY KEY,
			publisher_id INT NOT NULL,
			keyword_id INT,
			keyword_title VARCHAR(500),
			slot VARCHAR(100),
			client_ip VARCHAR(100),
			user_agent TEXT,
			country_code VARCHAR(10),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_publisher_id (publisher_id),
			INDEX idx_keyword_id (keyword_id),
			INDEX idx_created_at (created_at)
		)`,
		// Ad impression - records when ads are shown on SERP page
		`CREATE TABLE IF NOT EXISTS ad_impression (
			id INT AUTO_INCREMENT PRIMARY KEY,
			publisher_id INT NOT NULL,
			keyword_id INT,
			keyword_title VARCHAR(500),
			ad_position INT,
			ad_title VARCHAR(500),
			ad_host VARCHAR(255),
			client_ip VARCHAR(100),
			user_agent TEXT,
			country_code VARCHAR(10),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_publisher_id (publisher_id),
			INDEX idx_keyword_id (keyword_id),
			INDEX idx_created_at (created_at)
		)`,
		// Ad click - records when an ad is clicked on SERP page
		`CREATE TABLE IF NOT EXISTS ad_click (
			id INT AUTO_INCREMENT PRIMARY KEY,
			publisher_id INT NOT NULL,
			keyword_id INT,
			keyword_title VARCHAR(500),
			ad_title VARCHAR(500),
			ad_host VARCHAR(255),
			ad_target_url TEXT,
			slot VARCHAR(100),
			client_ip VARCHAR(100),
			user_agent TEXT,
			country_code VARCHAR(10),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_publisher_id (publisher_id),
			INDEX idx_keyword_id (keyword_id),
			INDEX idx_created_at (created_at)
		)`,
	}

	for _, query := range tables {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	log.Println("Ensured all tables exist")

	// Seed publishers
	if err := seedPublishers(); err != nil {
		return fmt.Errorf("failed to seed publishers: %w", err)
	}

	return nil
}

// seedPublishers seeds the publisher table with known publishers
func seedPublishers() error {
	publishers := []struct {
		publisherID int
		domain      string
	}{
		{100, "blue"},
		{200, "red"},
	}

	for _, pub := range publishers {
		_, err := DB.Exec(`
			INSERT INTO publisher (publisher_id, domain)
			VALUES (?, ?)
			ON DUPLICATE KEY UPDATE domain = VALUES(domain)
		`, pub.publisherID, pub.domain)
		if err != nil {
			return fmt.Errorf("failed to seed publisher %d: %w", pub.publisherID, err)
		}
	}

	log.Println("Publishers seeded: 100=blue, 200=red")
	return nil
}

// seedRules seeds default rule configurations
func seedRules() error {
	// Check if table already has records - skip seeding if not empty
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM rules").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check rules count: %w", err)
	}
	if count > 0 {
		log.Printf("Rules table has %d records, skipping seed", count)
		return nil
	}

	// Table is empty, seed default rules
	rules := []struct {
		ruleName    string
		action      string
		publisherID int
		userAgent   string
		countryCode string
	}{
		{
			ruleName: "publisher_100_default",
			action: `{
				"serp_template_id": "SerpTemplate2.html",
				"keyword_template_id": "KeywordTemplate1.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": false
			}`,
			publisherID: 100,
			userAgent:   "",
			countryCode: "US",
		},
		{
			ruleName: "publisher_100_android",
			action: `{
				"serp_template_id": "SerpTemplate2.html",
				"keyword_template_id": "KeywordTemplate1.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": true
			}`,
			publisherID: 100,
			userAgent:   "Android",
			countryCode: "US",
		},
		{
			ruleName: "publisher_100_googlebot",
			action: `{
				"serp_template_id": "SerptemplateBot.html",
				"keyword_template_id": "KeywordTemplateBot.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": true
			}`,
			publisherID: 100,
			userAgent:   "Googlebot",
			countryCode: "US",
		},
		{
			ruleName: "publisher_100_bot",
			action: `{
				"serp_template_id": "SerptemplateBot.html",
				"keyword_template_id": "KeywordTemplateBot.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": false
			}`,
			publisherID: 100,
			userAgent:   "bot",
			countryCode: "US",
		},
		{
			ruleName: "publisher_200_default",
			action: `{
				"serp_template_id": "SerpTemplate3.html",
				"keyword_template_id": "KeywordTemplate3.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": false
			}`,
			publisherID: 200,
			userAgent:   "",
			countryCode: "US",
		},
		{
			ruleName: "publisher_200_android",
			action: `{
				"serp_template_id": "SerpTemplate3.html",
				"keyword_template_id": "KeywordTemplate3.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": true
			}`,
			publisherID: 200,
			userAgent:   "Android",
			countryCode: "US",
		},
		{
			ruleName: "publisher_200_googlebot",
			action: `{
				"serp_template_id": "SerptemplateBot.html",
				"keyword_template_id": "KeywordTemplateBot.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": true
			}`,
			publisherID: 200,
			userAgent:   "Googlebot",
			countryCode: "US",
		},
		{
			ruleName: "publisher_200_bot",
			action: `{
				"serp_template_id": "SerptemplateBot.html",
				"keyword_template_id": "KeywordTemplateBot.html",
				"layout_id": 224,
				"template_size": "300x250",
				"block": false,
				"open_in_new_tab": false
			}`,
			publisherID: 200,
			userAgent:   "bot",
			countryCode: "US",
		},
	}

	for _, rule := range rules {
		_, err := DB.Exec(`
			INSERT INTO rules (rule_name, action, publisher_id, user_agent, country_code)
			VALUES (?, ?, ?, ?, ?)
		`, rule.ruleName, rule.action, rule.publisherID, rule.userAgent, rule.countryCode)
		if err != nil {
			return fmt.Errorf("failed to seed rule %s: %w", rule.ruleName, err)
		}
		log.Printf("Seeded rule: %s for publisher_id %d", rule.ruleName, rule.publisherID)
	}

	log.Println("Rules seeding complete")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return DB
}
