package config

import (
	"database/sql"
	"encoding/json"
	"log"
)

type RuleAction struct {
	SerpTemplateID    string `json:"serp_template_id"`
	KeywordTemplateID string `json:"keyword_template_id"`
	Block             bool   `json:"block"`
	OpenInNewTab      bool   `json:"open_in_new_tab"`
}

type Rule struct {
	ID          int
	RuleName    string
	Action      RuleAction
	PublisherID int
	UserAgent   string
	CountryCode string
}

var DefaultRuleAction = RuleAction{
	SerpTemplateID:    "SerpTemplate1.html",
	KeywordTemplateID: "KeywordTemplate1.html",
	Block:             false,
	OpenInNewTab:      false,
}

var DefaultRule = Rule{
	RuleName:    "default",
	Action:      DefaultRuleAction,
	CountryCode: "US",
}

var rulesDBConn *sql.DB

func SetRulesDB(db *sql.DB) {
	rulesDBConn = db
}

func GetRuleByPublisherID(publisherID int) Rule {
	if publisherID == 0 || rulesDBConn == nil {
		return DefaultRule
	}

	var rule Rule
	var actionJSON string
	err := rulesDBConn.QueryRow(`
		SELECT id, rule_name, action, publisher_id, COALESCE(user_agent, ''), COALESCE(country_code, 'US')
		FROM rules WHERE publisher_id = ? AND (user_agent IS NULL OR user_agent = '') LIMIT 1
	`, publisherID).Scan(&rule.ID, &rule.RuleName, &actionJSON, &rule.PublisherID, &rule.UserAgent, &rule.CountryCode)

	if err != nil {
		return DefaultRule
	}

	if err := json.Unmarshal([]byte(actionJSON), &rule.Action); err != nil {
		log.Printf("action JSON parse error: %v", err)
		rule.Action = DefaultRuleAction
	}
	return rule
}

func GetRuleByPublisherIDAndUserAgent(publisherID int, userAgent string) Rule {
	if publisherID == 0 || rulesDBConn == nil {
		return DefaultRule
	}

	var rule Rule
	var actionJSON string
	err := rulesDBConn.QueryRow(`
		SELECT id, rule_name, action, publisher_id, COALESCE(user_agent, ''), COALESCE(country_code, 'US')
		FROM rules WHERE publisher_id = ? AND user_agent != '' AND ? LIKE CONCAT('%', user_agent, '%')
		ORDER BY LENGTH(user_agent) DESC LIMIT 1
	`, publisherID, userAgent).Scan(&rule.ID, &rule.RuleName, &actionJSON, &rule.PublisherID, &rule.UserAgent, &rule.CountryCode)

	if err != nil {
		return GetRuleByPublisherID(publisherID)
	}

	if err := json.Unmarshal([]byte(actionJSON), &rule.Action); err != nil {
		log.Printf("action JSON parse error: %v", err)
		rule.Action = DefaultRuleAction
	}
	return rule
}

func UpsertRule(rule Rule) error {
	if rulesDBConn == nil {
		return nil
	}

	actionJSON, err := json.Marshal(rule.Action)
	if err != nil {
		return err
	}

	_, err = rulesDBConn.Exec(`
		INSERT INTO rules (rule_name, action, publisher_id, user_agent, country_code) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE rule_name = VALUES(rule_name), action = VALUES(action), user_agent = VALUES(user_agent), country_code = VALUES(country_code)
	`, rule.RuleName, string(actionJSON), rule.PublisherID, rule.UserAgent, rule.CountryCode)

	return err
}
