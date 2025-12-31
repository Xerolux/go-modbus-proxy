# Debian/Ubuntu Installation Guide

## Schnellinstallation mit .deb Paket

Die einfachste Methode zur Installation von ModBridge auf Debian oder Ubuntu.

---

## Unterstützte Systeme

### Debian
- Debian 11 (Bullseye)
- Debian 12 (Bookworm)
- Debian Testing/Unstable

### Ubuntu
- Ubuntu 20.04 LTS (Focal Fossa)
- Ubuntu 22.04 LTS (Jammy Jellyfish)
- Ubuntu 24.04 LTS (Noble Numbat)

### Andere
- Raspberry Pi OS (64-bit)
- Linux Mint
- Pop!_OS
- Alle anderen Debian-basierten Distributionen mit systemd

---

## Verfügbare Pakete

### AMD64 (x86_64)
- **Datei**: `modbus-proxy-manager_0.1.0_amd64.deb`
- **Größe**: ~2.4 MB
- **Für**: Desktop-PCs, Server, Intel/AMD-Prozessoren

### ARM64 (aarch64)
- **Datei**: `modbus-proxy-manager_0.1.0_arm64.deb`
- **Größe**: ~2.1 MB
- **Für**: Raspberry Pi 4/5 (64-bit), ARM-Server

---

## Installation

### Schritt 1: Paket herunterladen

```bash
# Für AMD64 (x86_64)
wget https://github.com/Xerolux/modbridge/releases/download/v0.1.0/modbus-proxy-manager_0.1.0_amd64.deb

# Für ARM64 (Raspberry Pi 64-bit)
wget https://github.com/Xerolux/modbridge/releases/download/v0.1.0/modbus-proxy-manager_0.1.0_arm64.deb
```

**Oder** klone das Repository:

```bash
git clone https://github.com/Xerolux/modbridge.git
cd modbridge/releases
```

### Schritt 2: Paket installieren

```bash
# Für AMD64
sudo dpkg -i modbus-proxy-manager_0.1.0_amd64.deb

# Für ARM64
sudo dpkg -i modbus-proxy-manager_0.1.0_arm64.deb
```

**Bei fehlenden Abhängigkeiten**:

```bash
sudo apt-get install -f
```

### Schritt 3: Service starten

```bash
sudo systemctl start modbusmanager
```

### Schritt 4: Status überprüfen

```bash
sudo systemctl status modbusmanager
```

### Schritt 5: Web-Interface öffnen

Öffnen Sie Ihren Browser:

```
http://localhost:8080
```

Oder von einem anderen Rechner im Netzwerk:

```
http://<SERVER-IP>:8080
```

---

## Was wird installiert?

### Dateien und Verzeichnisse

```
/opt/modbusmanager/
├── modbusmanager                    # Das Haupt-Binary

/var/lib/modbusmanager/
├── config.json                      # Konfigurationsdatei

/var/log/modbusmanager/
└── (Log-Dateien)                    # System-Logs

/etc/systemd/system/
└── modbusmanager.service            # Systemd Service
```

### System-User

- **User**: `modbusmanager`
- **Home**: Kein Home-Verzeichnis
- **Shell**: `/bin/false` (kein Login möglich)
- **Zweck**: Sicherheit (Least Privilege Principle)

---

## Service-Verwaltung

### Status prüfen

```bash
sudo systemctl status modbusmanager
```

### Service starten

```bash
sudo systemctl start modbusmanager
```

### Service stoppen

```bash
sudo systemctl stop modbusmanager
```

### Service neustarten

```bash
sudo systemctl restart modbusmanager
```

### Autostart aktivieren

```bash
sudo systemctl enable modbusmanager
```

Der Service wird **automatisch** beim Systemstart aktiviert, wenn das Paket installiert wird.

### Autostart deaktivieren

```bash
sudo systemctl disable modbusmanager
```

---

## Logs ansehen

### Live-Logs (folgt neuen Einträgen)

```bash
sudo journalctl -u modbusmanager -f
```

### Letzte 100 Log-Zeilen

```bash
sudo journalctl -u modbusmanager -n 100
```

### Logs seit heute

```bash
sudo journalctl -u modbusmanager --since today
```

### Logs der letzten Stunde

```bash
sudo journalctl -u modbusmanager --since "1 hour ago"
```

---

## Konfiguration

### Konfigurationsdatei bearbeiten

```bash
sudo nano /var/lib/modbusmanager/config.json
```

**Standard-Inhalt**:

```json
{
  "web_port": ":8080",
  "admin_pass_hash": "",
  "proxies": []
}
```

### Nach Änderungen Service neustarten

```bash
sudo systemctl restart modbusmanager
```

### Web-Port ändern

1. Datei bearbeiten:

```bash
sudo nano /etc/systemd/system/modbusmanager.service
```

2. Zeile ändern:

```ini
Environment="WEB_PORT=:9090"
```

3. Systemd neu laden und Service neustarten:

```bash
sudo systemctl daemon-reload
sudo systemctl restart modbusmanager
```

---

## Update / Upgrade

### Auf neue Version aktualisieren

1. **Neue Version herunterladen**:

```bash
wget https://github.com/Xerolux/modbridge/releases/download/v0.2.0/modbus-proxy-manager_0.2.0_amd64.deb
```

2. **Paket installieren** (überschreibt alte Version):

```bash
sudo dpkg -i modbus-proxy-manager_0.2.0_amd64.deb
```

3. **Service wird automatisch neugestartet**

**Hinweis**: Ihre Konfiguration (`config.json`) wird **nicht überschrieben**.

---

## Deinstallation

### Service beenden und Paket entfernen

```bash
sudo apt-get remove modbus-proxy-manager
```

**Konfiguration und Daten bleiben erhalten**.

### Vollständiges Entfernen (inkl. Konfiguration)

```bash
sudo apt-get purge modbus-proxy-manager
```

**Entfernt**:
- Binary (`/opt/modbusmanager/`)
- Konfiguration (`/var/lib/modbusmanager/`)
- Logs (`/var/log/modbusmanager/`)
- System-User (`modbusmanager`)
- Systemd Service

---

## Troubleshooting

### Port bereits in Verwendung

**Fehler**:
```
bind: address already in use
```

**Lösung**:

```bash
# Prozess finden
sudo lsof -i :8080

# Prozess beenden
sudo kill -9 <PID>

# Oder anderen Port verwenden (siehe "Web-Port ändern")
```

### Service startet nicht

**Diagnose**:

```bash
# Status prüfen
sudo systemctl status modbusmanager

# Detaillierte Logs
sudo journalctl -u modbusmanager -n 50
```

**Häufige Ursachen**:
- Port bereits belegt
- Fehlerhafte `config.json` Syntax
- Fehlende Berechtigungen

### Keine Verbindung zum Modbus-Gerät

**Test**:

```bash
# Ping-Test
ping 192.168.1.100

# Port-Test
nc -zv 192.168.1.100 502
```

**Prüfen**:
- Firewall-Regeln
- Netzwerk-Routing
- Zielgerät erreichbar?

### Binary funktioniert nicht

**Falsche Architektur?**

```bash
# System-Architektur prüfen
uname -m

# x86_64 = AMD64-Paket benötigt
# aarch64 = ARM64-Paket benötigt
```

**Abhängigkeiten installieren**:

```bash
sudo apt-get install -f
```

### Permission Denied

**Berechtigungen prüfen**:

```bash
ls -la /opt/modbusmanager/modbusmanager
ls -la /var/lib/modbusmanager/
```

**Korrigieren**:

```bash
sudo chown -R modbusmanager:modbusmanager /opt/modbusmanager
sudo chown -R modbusmanager:modbusmanager /var/lib/modbusmanager
sudo chown -R modbusmanager:modbusmanager /var/log/modbusmanager
```

---

## Performance-Optimierung

### Für hohe Last (1000+ Verbindungen)

1. **File Descriptor Limits erhöhen**:

```bash
sudo nano /etc/systemd/system/modbusmanager.service
```

Hinzufügen:

```ini
LimitNOFILE=65536
```

2. **Service neu laden**:

```bash
sudo systemctl daemon-reload
sudo systemctl restart modbusmanager
```

### TCP-Parameter optimieren (System-weit)

```bash
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
sudo sysctl -w net.ipv4.tcp_fin_timeout=30
sudo sysctl -w net.core.somaxconn=1024
```

**Permanent machen**:

```bash
sudo nano /etc/sysctl.conf
```

Hinzufügen:

```
net.ipv4.tcp_tw_reuse=1
net.ipv4.tcp_fin_timeout=30
net.core.somaxconn=1024
```

Anwenden:

```bash
sudo sysctl -p
```

---

## Firewall-Konfiguration

### UFW (Ubuntu Firewall)

```bash
# Web-Interface (Port 8080)
sudo ufw allow 8080/tcp

# Modbus Proxy-Ports (5020-5030)
sudo ufw allow 5020:5030/tcp

# Firewall neu laden
sudo ufw reload
```

### iptables

```bash
# Web-Interface
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# Modbus Proxy-Ports
sudo iptables -A INPUT -p tcp --dport 5020:5030 -j ACCEPT

# Speichern (Debian/Ubuntu)
sudo netfilter-persistent save
```

---

## Automatisches Backup

### Konfiguration sichern

```bash
# Backup erstellen
sudo cp /var/lib/modbusmanager/config.json \
       /var/lib/modbusmanager/config.json.backup

# Mit Datum
sudo cp /var/lib/modbusmanager/config.json \
       /var/lib/modbusmanager/config.json.$(date +%Y%m%d)
```

### Cron-Job für tägliches Backup

```bash
sudo crontab -e
```

Hinzufügen:

```cron
# Täglich um 2:00 Uhr
0 2 * * * cp /var/lib/modbusmanager/config.json /var/lib/modbusmanager/config.json.$(date +\%Y\%m\%d)

# Alte Backups löschen (älter als 30 Tage)
0 3 * * * find /var/lib/modbusmanager/ -name "config.json.*" -mtime +30 -delete
```

---

## Sicherheits-Tipps

### 1. Admin-Passwort setzen

Beim ersten Zugriff auf `http://localhost:8080` werden Sie aufgefordert, ein Admin-Passwort zu setzen.

**Empfehlung**: Mindestens 12 Zeichen, Groß-/Kleinbuchstaben, Zahlen, Sonderzeichen.

### 2. Nur auf localhost lauschen

Wenn Sie nur lokal zugreifen:

```bash
sudo nano /etc/systemd/system/modbusmanager.service
```

Ändern:

```ini
Environment="WEB_PORT=127.0.0.1:8080"
```

```bash
sudo systemctl daemon-reload
sudo systemctl restart modbusmanager
```

### 3. Reverse Proxy mit SSL

Für Produktion empfohlen: Nginx/Apache mit SSL (HTTPS).

**Beispiel Nginx**:

```nginx
server {
    listen 443 ssl;
    server_name modbus.example.com;

    ssl_certificate /etc/ssl/certs/modbus.crt;
    ssl_certificate_key /etc/ssl/private/modbus.key;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Support

### Logs überprüfen

```bash
sudo journalctl -u modbusmanager -n 100
```

### GitHub Issue erstellen

https://github.com/Xerolux/modbridge/issues

**Bitte angeben**:
- Linux-Distribution und Version (`lsb_release -a`)
- Paket-Version (`dpkg -l | grep modbus-proxy-manager`)
- Logs (`journalctl -u modbusmanager -n 100`)
- Konfiguration (ohne Passwort-Hash!)
- Fehlermeldung

---

## Nützliche Befehle

```bash
# Paket-Informationen anzeigen
dpkg -l | grep modbus-proxy-manager

# Installierte Dateien auflisten
dpkg -L modbus-proxy-manager

# Paket-Details
apt show modbus-proxy-manager

# Konfiguration exportieren (Backup)
sudo cp /var/lib/modbusmanager/config.json ~/config-backup.json

# Konfiguration wiederherstellen
sudo cp ~/config-backup.json /var/lib/modbusmanager/config.json
sudo systemctl restart modbusmanager

# Service komplett neu installieren
sudo apt-get purge modbus-proxy-manager
sudo dpkg -i modbus-proxy-manager_0.1.0_amd64.deb
```

---

## Weitere Ressourcen

- **Haupt-README**: [README.md](README.md)
- **Performance-Guide**: [docs/PERFORMANCE.md](docs/PERFORMANCE.md)
- **Roadmap**: [ROADMAP.md](ROADMAP.md)
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md)

---

**Version**: 0.1.0
**Letzte Aktualisierung**: 31. Dezember 2025
**Status**: Production Ready
