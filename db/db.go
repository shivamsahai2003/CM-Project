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

	// Seed default publisher configs
	if err := seedPublisherConfigs(); err != nil {
		return fmt.Errorf("failed to seed publisher configs: %w", err)
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
		`CREATE TABLE IF NOT EXISTS publisher (
			publisher_id INT PRIMARY KEY,
			domain VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS pub_config (
			pid INT PRIMARY KEY,
			lid INT DEFAULT 224,
			cc VARCHAR(10) DEFAULT 'US',
			tsize VARCHAR(20) DEFAULT '300x250',
			serp_template VARCHAR(255) DEFAULT 'SerpTemplate1.html',
			keyword_template VARCHAR(255) DEFAULT 'KeywordTemplate1.html',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS keyword_impression (
			id INT AUTO_INCREMENT PRIMARY KEY,
			publisher_id INT,
			keyword_no INT,
			keywords VARCHAR(500),
			slot VARCHAR(100),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS adclick_click (
			id INT AUTO_INCREMENT PRIMARY KEY,
			keyword_id INT,
			time DATETIME,
			` + "`user id`" + ` VARCHAR(100),
			keyword_title VARCHAR(500),
			Ad_details TEXT,
			User_agent TEXT,
			publisher_id INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS keyword_click (
			id INT AUTO_INCREMENT PRIMARY KEY,
			slot_id INT,
			kid INT,
			time DATETIME,
			` + "`user id`" + ` VARCHAR(100),
			keyword_title VARCHAR(500),
			publisher_id INT,
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range tables {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	log.Println("Ensured all tables exist")
	return nil
}

// seedPublisherConfigs seeds default publisher configurations
func seedPublisherConfigs() error {
	// Check if table already has records - skip seeding if not empty
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM pub_config").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check pub_config count: %w", err)
	}
	if count > 0 {
		log.Printf("Publisher config table has %d records, skipping seed", count)
		return nil
	}

	// Table is empty, seed default configs
	configs := []struct {
		pid             int
		lid             int
		cc              string
		tsize           string
		serpTemplate    string
		keywordTemplate string
	}{
		{
			pid:             100,
			lid:             224,
			cc:              "US",
			tsize:           "300x250",
			serpTemplate:    "SerpTemplate2.html",    // 2 ads (template decides)
			keywordTemplate: "KeywordTemplate1.html", // 6 keywords (template decides)
		},
		{
			pid:             200,
			lid:             224,
			cc:              "US",
			tsize:           "300x250",
			serpTemplate:    "SerpTemplate3.html",    // 5 ads (template decides)
			keywordTemplate: "KeywordTemplate3.html", // 10 keywords (template decides)
		},
	}

	for _, cfg := range configs {
		_, err := DB.Exec(`
			INSERT INTO pub_config (pid, lid, cc, tsize, serp_template, keyword_template)
			VALUES (?, ?, ?, ?, ?, ?)
		`, cfg.pid, cfg.lid, cfg.cc, cfg.tsize, cfg.serpTemplate, cfg.keywordTemplate)
		if err != nil {
			return fmt.Errorf("failed to seed config for PID %d: %w", cfg.pid, err)
		}
		log.Printf("Seeded publisher config for PID %d", cfg.pid)
	}

	log.Println("Publisher config seeding complete")
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
