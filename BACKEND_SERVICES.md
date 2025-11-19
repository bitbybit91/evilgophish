# Backend Services Configuration Guide

EvilGoPhish Hidden Services now supports multiple backend services running behind your Tor v3 hidden service. This guide explains how to configure each type.

## Supported Backend Services

### 1. EvilGinx3 (Default)

**Description:** MITM proxy for credential harvesting with 2FA/MFA bypass capabilities.

**Installation:**
```bash
sudo ./setup_hidden_services.sh user_id evilginx3 443
```

**Default Port:** 443  
**Use Case:** Advanced phishing campaigns with session token capture

**Post-Installation:**
- Access GoPhish at `https://localhost:3333`
- Configure phishlets in evilginx3
- Create lures for your target services
- Use your .onion address in phishing campaigns

---

### 2. PHP Applications

**Description:** Generic PHP application with PHP-FPM support for static websites or PHP frameworks.

**Installation:**
```bash
sudo ./setup_hidden_services.sh user_id php 80
```

**Default Port:** 80  
**Use Case:** Static HTML sites with PHP, Laravel, Symfony, or custom PHP applications

**Post-Installation:**
1. Place your PHP files in `/var/www/html/`
2. Configure PHP-FPM settings if needed:
   ```bash
   sudo nano /etc/php/*/fpm/pool.d/www.conf
   ```
3. Restart PHP-FPM:
   ```bash
   sudo systemctl restart php*-fpm
   ```
4. Your site will be accessible via your .onion address

**Example - Deploy Static PHP Site:**
```bash
# Copy your PHP files
sudo cp -r /path/to/your/site/* /var/www/html/

# Set permissions
sudo chown -R www-data:www-data /var/www/html
sudo chmod -R 755 /var/www/html

# Test locally
curl http://localhost/
```

**Supported PHP Frameworks:**
- Laravel
- Symfony
- CodeIgniter
- CakePHP
- Slim
- Yii
- Zend Framework
- Custom PHP applications

---

### 3. WordPress CMS

**Description:** Complete WordPress installation with MySQL database for realistic phishing pages.

**Installation:**
```bash
sudo ./setup_hidden_services.sh user_id wordpress 80
```

**Default Port:** 80  
**Use Case:** Create realistic login pages that look like legitimate WordPress sites

**Post-Installation:**

1. **Secure MySQL:**
   ```bash
   sudo mysql_secure_installation
   ```

2. **Create WordPress Database:**
   ```bash
   sudo mysql -u root -p
   ```
   ```sql
   CREATE DATABASE wordpress;
   CREATE USER 'wpuser'@'localhost' IDENTIFIED BY 'strong_password';
   GRANT ALL PRIVILEGES ON wordpress.* TO 'wpuser'@'localhost';
   FLUSH PRIVILEGES;
   EXIT;
   ```

3. **Download and Install WordPress:**
   ```bash
   cd /tmp
   wget https://wordpress.org/latest.tar.gz
   tar -xvf latest.tar.gz
   sudo cp -r wordpress/* /var/www/html/
   sudo chown -R www-data:www-data /var/www/html
   sudo chmod -R 755 /var/www/html
   ```

4. **Configure WordPress:**
   ```bash
   cd /var/www/html
   sudo cp wp-config-sample.php wp-config.php
   sudo nano wp-config.php
   ```
   
   Update database settings:
   ```php
   define('DB_NAME', 'wordpress');
   define('DB_USER', 'wpuser');
   define('DB_PASSWORD', 'strong_password');
   define('DB_HOST', 'localhost');
   ```

5. **Complete Installation:**
   - Access via .onion address
   - Follow WordPress installation wizard
   - Install themes and plugins as needed

**Tips for Phishing Pages:**
- Clone legitimate login pages with themes
- Use similar-looking themes to target sites
- Customize login forms to capture credentials
- Integrate with GoPhish for campaign tracking

---

### 4. Apache Web Server

**Description:** Traditional Apache HTTP server for serving static or dynamic content.

**Installation:**
```bash
sudo ./setup_hidden_services.sh user_id apache 80
```

**Default Port:** 80  
**Use Case:** Static websites, CGI scripts, or when you need Apache-specific features

**Post-Installation:**

1. **Deploy Your Site:**
   ```bash
   sudo cp -r /path/to/your/site/* /var/www/html/
   sudo chown -R www-data:www-data /var/www/html
   ```

2. **Enable Modules (if needed):**
   ```bash
   sudo a2enmod rewrite ssl headers
   sudo systemctl restart apache2
   ```

3. **Configure Virtual Hosts (optional):**
   ```bash
   sudo nano /etc/apache2/sites-available/000-default.conf
   ```

4. **Check Configuration:**
   ```bash
   sudo apache2ctl configtest
   sudo systemctl restart apache2
   ```

**Example - Enable .htaccess:**
```apache
<Directory /var/www/html>
    Options Indexes FollowSymLinks
    AllowOverride All
    Require all granted
</Directory>
```

---

### 5. Nginx Web Server

**Description:** High-performance web server ideal for serving static content or as a reverse proxy.

**Installation:**
```bash
sudo ./setup_hidden_services.sh user_id nginx 80
```

**Default Port:** 80  
**Use Case:** High-traffic static sites, single-page applications, or reverse proxy scenarios

**Post-Installation:**

1. **Deploy Your Site:**
   ```bash
   sudo cp -r /path/to/your/site/* /var/www/html/
   sudo chown -R www-data:www-data /var/www/html
   ```

2. **Configure Nginx:**
   ```bash
   sudo nano /etc/nginx/sites-available/default
   ```

3. **Basic Configuration Example:**
   ```nginx
   server {
       listen 80 default_server;
       root /var/www/html;
       index index.html index.htm index.php;
       
       location / {
           try_files $uri $uri/ =404;
       }
       
       # PHP support (if needed)
       location ~ \.php$ {
           include snippets/fastcgi-php.conf;
           fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
       }
   }
   ```

4. **Test and Reload:**
   ```bash
   sudo nginx -t
   sudo systemctl reload nginx
   ```

---

### 6. Custom Backend Service

**Description:** Any HTTP/HTTPS service you want to proxy through Tor.

**Installation:**
```bash
sudo ./setup_hidden_services.sh user_id custom 3000
```

**Port:** Specify your service's port  
**Use Case:** Node.js apps, Python Flask/Django, Ruby on Rails, Go web servers, etc.

**Examples:**

**Node.js Express App:**
```bash
# Your app runs on port 3000
sudo ./setup_hidden_services.sh user_id custom 3000

# Start your Node.js app
cd /path/to/your/app
npm start
```

**Python Flask:**
```bash
# Flask runs on port 5000
sudo ./setup_hidden_services.sh user_id custom 5000

# Start Flask
export FLASK_APP=app.py
flask run --host=127.0.0.1 --port=5000
```

**Go Web Server:**
```bash
# Go server on port 8080
sudo ./setup_hidden_services.sh user_id custom 8080

# Start your Go app
./your-go-binary
```

**Docker Container:**
```bash
# Docker container exposing port 8888
sudo ./setup_hidden_services.sh user_id custom 8888

# Run your container
docker run -p 127.0.0.1:8888:80 your-image
```

---

## Backend Configuration Summary

| Backend | Default Port | Install Command | Use Case |
|---------|-------------|-----------------|----------|
| EvilGinx3 | 443 | `user_id evilginx3 443` | MITM credential harvesting |
| PHP | 80 | `user_id php 80` | PHP frameworks/applications |
| WordPress | 80 | `user_id wordpress 80` | CMS phishing pages |
| Apache | 80 | `user_id apache 80` | Static/dynamic websites |
| Nginx | 80 | `user_id nginx 80` | High-performance sites |
| Custom | Your choice | `user_id custom PORT` | Any HTTP service |

---

## How Tor Proxying Works

```
Victim Browser
    ↓
Tor Network (.onion address)
    ↓
Your VPS (Tor Hidden Service)
    ↓
Port 80/443 forwarded to →
    ↓
Backend Service (127.0.0.1:YOUR_PORT)
    ↓
Your Application (EvilGinx3/PHP/WordPress/etc.)
```

**Configuration in `/etc/tor/torrc`:**
```
HiddenServiceDir /var/lib/tor/evilgophish/
HiddenServiceVersion 3
HiddenServicePort 80 127.0.0.1:YOUR_BACKEND_PORT
HiddenServicePort 443 127.0.0.1:YOUR_BACKEND_PORT
```

---

## Multiple Backend Instances

You can run multiple backends on different ports:

```bash
# EvilGinx3 on 443
sudo ./setup_hidden_services.sh user_id evilginx3 443

# Also run WordPress on 8080
# (Requires manual Tor configuration for additional service)
```

**For multiple hidden services, manually edit `/etc/tor/torrc`:**
```
# First hidden service (EvilGinx3)
HiddenServiceDir /var/lib/tor/evilgophish/
HiddenServiceVersion 3
HiddenServicePort 80 127.0.0.1:443
HiddenServicePort 443 127.0.0.1:443

# Second hidden service (WordPress)
HiddenServiceDir /var/lib/tor/evilgophish_wp/
HiddenServiceVersion 3
HiddenServicePort 80 127.0.0.1:8080
HiddenServicePort 443 127.0.0.1:8080
```

Then restart Tor:
```bash
sudo systemctl restart tor
cat /var/lib/tor/evilgophish/hostname      # First .onion
cat /var/lib/tor/evilgophish_wp/hostname   # Second .onion
```

---

## Security Best Practices

### For All Backends:

1. **Bind to localhost only:**
   - Ensure services only listen on `127.0.0.1`
   - Never bind to `0.0.0.0` or public IPs

2. **Firewall rules:**
   ```bash
   sudo ufw default deny incoming
   sudo ufw default allow outgoing
   sudo ufw allow ssh
   sudo ufw enable
   ```

3. **Regular updates:**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

4. **Log monitoring:**
   ```bash
   sudo journalctl -u tor -f
   tail -f /var/log/nginx/access.log  # or apache2
   ```

### PHP/WordPress Specific:

1. **Disable dangerous functions:**
   ```ini
   # /etc/php/*/fpm/php.ini
   disable_functions = exec,passthru,shell_exec,system,proc_open,popen
   ```

2. **Hide PHP version:**
   ```ini
   expose_php = Off
   ```

3. **File permissions:**
   ```bash
   find /var/www/html -type d -exec chmod 755 {} \;
   find /var/www/html -type f -exec chmod 644 {} \;
   ```

---

## Troubleshooting

### Backend not accessible via .onion:

1. **Check backend is running:**
   ```bash
   curl http://localhost:YOUR_PORT
   ```

2. **Check Tor configuration:**
   ```bash
   sudo cat /etc/tor/torrc | grep -A5 HiddenService
   ```

3. **Check Tor logs:**
   ```bash
   sudo journalctl -u tor -n 50
   ```

4. **Verify port mapping:**
   ```bash
   sudo netstat -tulpn | grep YOUR_PORT
   ```

### PHP/WordPress issues:

1. **Check PHP-FPM:**
   ```bash
   sudo systemctl status php*-fpm
   ```

2. **Check PHP error logs:**
   ```bash
   sudo tail -f /var/log/php*-fpm.log
   ```

3. **Test PHP:**
   ```bash
   echo "<?php phpinfo(); ?>" | sudo tee /var/www/html/info.php
   curl http://localhost/info.php
   ```

---

## Examples and Use Cases

### Use Case 1: Corporate Portal Clone (WordPress)

**Scenario:** Clone a corporate login portal using WordPress

```bash
# Install with WordPress backend
sudo ./setup_hidden_services.sh user_id wordpress 80

# Install a theme that matches target
# Configure form to capture credentials
# Use .onion address in phishing email
```

### Use Case 2: Custom PHP Credential Harvester

**Scenario:** Custom-built PHP login form

```bash
# Install with PHP backend
sudo ./setup_hidden_services.sh user_id php 80

# Deploy your custom PHP login form
sudo cp custom_login.php /var/www/html/index.php

# Credentials POST to credential capture script
# Integrates with report_monitor for Telegram alerts
```

### Use Case 3: Static HTML Phishing Page (Nginx)

**Scenario:** High-performance static HTML/CSS/JS page

```bash
# Install with Nginx backend
sudo ./setup_hidden_services.sh user_id nginx 80

# Deploy your static site
sudo cp -r static_site/* /var/www/html/

# Fast serving for high traffic campaigns
```

---

## Additional Resources

- [HIDDEN_SERVICES_README.md](HIDDEN_SERVICES_README.md) - Complete setup guide
- [QUICKSTART.md](QUICKSTART.md) - 5-minute quick start
- [Tor v3 Protocol Specification](https://gitweb.torproject.org/torspec.git/tree/proposals/224-rend-spec-ng.txt)
- [PHP-FPM Configuration](https://www.php.net/manual/en/install.fpm.configuration.php)
- [WordPress Security Guide](https://wordpress.org/support/article/hardening-wordpress/)

---

**Remember:** Only use for authorized security testing with written permission!
