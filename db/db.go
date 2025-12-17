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
