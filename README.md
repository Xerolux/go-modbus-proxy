# ModBridge - Modbus TCP Proxy Manager

**Version:** 0.1.0

[![CI/CD Pipeline](https://github.com/Xerolux/modbridge/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/Xerolux/modbridge/actions/workflows/main.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/Xerolux/modbridge/pkgs/container/modbridge)

Ein moderner, robuster Modbus TCP Proxy Manager mit einer eleganten Web-Oberfläche. ModBridge ermöglicht die Verwaltung mehrerer Modbus TCP Proxy-Instanzen über eine zentrale Webschnittstelle.

## 📦 Quick Install

```bash
# Debian/Ubuntu (.deb)
wget https://github.com/Xerolux/modbridge/releases/latest/download/modbridge_0.1.0_amd64.deb
sudo dpkg -i modbridge_0.1.0_amd64.deb

# Docker (vorgefertigt)
docker run -d -p 8080:8080 -p 5020-5030:5020-5030 ghcr.io/xerolux/modbridge:latest

# Docker Compose
docker-compose up -d
```

## Inhaltsverzeichnis

- [Features](#features)
- [Systemanforderungen](#systemanforderungen)
- [Installation](#installation)
  - [Methode 1: Debian/Ubuntu Paket (.deb)](#methode-1-debianubuntu-paket-deb---einfachste-installation)
  - [Methode 2: Docker](#methode-2-docker)
  - [Methode 3: Aus Quellcode kompilieren](#methode-3-aus-quellcode-kompilieren)
  - [Methode 4: Systemd Service Installation](#methode-4-systemd-service-installation)
  - [Methode 5: Mit Makefile](#methode-5-mit-makefile)
- [Konfiguration](#konfiguration)
- [Verwendung](#verwendung)
- [Umgebungsvariablen](#umgebungsvariablen)
- [Performance](#performance)
- [Troubleshooting](#troubleshooting)
- [Lizenz](#lizenz)

---

## Features

### Kern-Funktionen
- **Multi-Proxy-Unterstützung**: Verwaltung mehrerer Modbus TCP Proxy-Instanzen
- **Web-Interface**: Moderne, responsive UI (eingebettet im Binary)
- **Echtzeit-Überwachung**: Live-Traffic-Logging via Server-Sent Events (SSE)
- **Sicherheit**:
  - Admin-Authentifizierung mit Bcrypt-Hashing
  - Sichere Session-Cookies
  - Optional: Read-Only-Ansicht für nicht authentifizierte Benutzer
- **Persistenz**: JSON-basierte Konfiguration mit Export/Import
- **Single Binary**: Gesamte Frontend ist im Go-Binary eingebettet
- **Graceful Shutdown**: Sauberes Herunterfahren aller Proxy-Instanzen
- **Health Checks**: Integrierte Gesundheitsprüfungen für Container-Umgebungen

### Performance-Merkmale
- Latenz (avg): ~3-5ms
- Latenz (p99): ~12ms
- Durchsatz: ~10,000 req/s
- Speicher (idle): ~8MB
- Gleichzeitige Verbindungen: ~1,000

---

## Systemanforderungen

### Minimale Anforderungen
- **CPU**: 1 Core (2+ empfohlen)
- **RAM**: 128MB (512MB empfohlen)
- **Festplatte**: 50MB
- **Netzwerk**: TCP-Ports (Standard: 8080 für Web-UI, 5020-5030 für Proxies)

### Software-Anforderungen

#### Für Quellcode-Kompilierung
- Go 1.24.0 oder höher
- Git

#### Für Docker-Installation
- Docker 20.10+
- Docker Compose 2.0+ (optional)

#### Für Systemd-Installation
- Linux mit systemd
- Root-Zugriff (sudo)

---

## Installation

### Methode 1: Debian/Ubuntu Paket (.deb) - Einfachste Installation

Die schnellste Methode für Debian und Ubuntu Systeme.

#### Voraussetzungen
- Debian 11+, Ubuntu 20.04+ oder andere Debian-basierte Distribution
- systemd

#### Installation

1. **Paket herunterladen**:
   ```bash
   # Für AMD64 (x86_64)
   wget https://github.com/Xerolux/modbridge/releases/download/v0.1.0/modbridge_0.1.0_amd64.deb

   # Für ARM64 (Raspberry Pi 64-bit)
   wget https://github.com/Xerolux/modbridge/releases/download/v0.1.0/modbridge_0.1.0_arm64.deb
   ```

2. **Paket installieren**:
   ```bash
   # Für AMD64
   sudo dpkg -i modbridge_0.1.0_amd64.deb

   # Für ARM64
   sudo dpkg -i modbridge_0.1.0_arm64.deb

   # Falls Abhängigkeiten fehlen
   sudo apt-get install -f
   ```

3. **Service starten**:
   ```bash
   sudo systemctl start modbridge
   ```

4. **Web-Interface öffnen**:
   ```
   http://localhost:8080
   ```

**Detaillierte Anleitung**: Siehe [INSTALL_DEBIAN.md](INSTALL_DEBIAN.md)

---

### Methode 2: Docker

Docker ist eine plattformunabhängige Installationsmethode.

#### Mit vorgebautem Image (GitHub Container Registry)

Das schnellste Docker-Setup - kein Build erforderlich!

1. **Image pullen und starten**:
   ```bash
   docker run -d \
     --name modbridge \
     -p 8080:8080 \
     -p 5020-5030:5020-5030 \
     -v $(pwd)/config.json:/app/config.json \
     -v $(pwd)/logs:/app/data \
     --restart unless-stopped \
     ghcr.io/xerolux/modbridge:latest
   ```

2. **Web-Interface öffnen**:
   ```
   http://localhost:8080
   ```

**Verfügbare Tags**:
- `latest` - Neueste stabile Version
- `v0.1.0` - Spezifische Version
- `main` - Neueste Entwicklungsversion
- `edge` - Bleeding Edge (main branch)

---

#### Mit Docker Compose

1. **Repository klonen**:
   ```bash
   git clone https://github.com/Xerolux/modbridge.git
   cd modbridge
   ```

2. **Docker Compose starten**:
   ```bash
   docker-compose up -d
   ```

3. **Web-Interface öffnen**:
   ```
   http://localhost:8080
   ```

4. **Logs ansehen**:
   ```bash
   docker-compose logs -f
   ```

5. **Container stoppen**:
   ```bash
   docker-compose down
   ```

#### Mit Docker (ohne Compose)

1. **Image bauen**:
   ```bash
   docker build -t modbridge .
   ```

2. **Container starten**:
   ```bash
   docker run -d \
     --name modbridge \
     -p 8080:8080 \
     -p 5020-5030:5020-5030 \
     -v $(pwd)/config.json:/app/config.json \
     -v $(pwd)/logs:/app/data \
     -e WEB_PORT=:8080 \
     modbridge
   ```

3. **Container verwalten**:
   ```bash
   # Status prüfen
   docker ps

   # Logs ansehen
   docker logs -f modbridge

   # Container stoppen
   docker stop modbridge

   # Container entfernen
   docker rm modbridge
   ```

---

### Methode 3: Aus Quellcode kompilieren

Für Entwicklung oder wenn Docker nicht verfügbar ist.

#### Voraussetzungen
- Go 1.24.0 oder höher muss installiert sein

#### Schnellstart

1. **Repository klonen**:
   ```bash
   git clone https://github.com/Xerolux/modbridge.git
   cd modbridge
   ```

2. **Abhängigkeiten laden**:
   ```bash
   go mod download
   ```

3. **Kompilieren**:
   ```bash
   go build -o modbridge main.go
   ```

4. **Ausführen**:
   ```bash
   ./modbridge
   ```

5. **Web-Interface öffnen**:
   ```
   http://localhost:8080
   ```

#### Für alle Plattformen kompilieren

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o modbridge-linux-amd64 main.go

# Linux ARM64 (Raspberry Pi 64-bit)
GOOS=linux GOARCH=arm64 go build -o modbridge-linux-arm64 main.go

# Linux ARM (Raspberry Pi 32-bit)
GOOS=linux GOARCH=arm go build -o modbridge-linux-arm main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o modbridge-windows.exe main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o modbridge-darwin-amd64 main.go

# macOS ARM (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o modbridge-darwin-arm64 main.go
```

#### Optimiertes Build (kleiner Binary)

```bash
go build -ldflags="-s -w" -o modbridge main.go
```

**Flags Erklärung**:
- `-s`: Entfernt Symbol-Tabelle
- `-w`: Entfernt DWARF-Debug-Informationen
- Ergebnis: ~50% kleineres Binary

---

### Methode 4: Systemd Service Installation

Für produktive Linux-Server mit systemd.

#### Automatische Installation

1. **Repository klonen**:
   ```bash
   git clone https://github.com/Xerolux/modbridge.git
   cd modbridge
   ```

2. **Installations-Script ausführen**:
   ```bash
   sudo ./install.sh
   ```

   Das Script führt automatisch aus:
   - Erkennung der System-Architektur
   - Kompilierung des Binaries (falls nicht vorhanden)
   - Erstellung eines System-Users (`modbridge`)
   - Installation nach `/opt/modbridge`
   - Einrichtung von Daten-Verzeichnissen
   - Systemd-Service-Konfiguration
   - Service-Start und Aktivierung

#### Manuelle Installation

1. **Binary kompilieren**:
   ```bash
   go build -ldflags="-s -w" -o modbridge main.go
   ```

2. **System-User erstellen**:
   ```bash
   sudo useradd --system --no-create-home --shell /bin/false modbridge
   ```

3. **Verzeichnisse erstellen**:
   ```bash
   sudo mkdir -p /opt/modbridge
   sudo mkdir -p /var/lib/modbridge
   sudo mkdir -p /var/log/modbridge
   ```

4. **Binary installieren**:
   ```bash
   sudo cp modbridge /opt/modbridge/
   sudo chmod +x /opt/modbridge/modbridge
   ```

5. **Berechtigungen setzen**:
   ```bash
   sudo chown -R modbridge:modbridge /opt/modbridge
   sudo chown -R modbridge:modbridge /var/lib/modbridge
   sudo chown -R modbridge:modbridge /var/log/modbridge
   ```

6. **Systemd Service kopieren**:
   ```bash
   sudo cp modbridge.service /etc/systemd/system/
   ```

7. **Service aktivieren und starten**:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable modbridge
   sudo systemctl start modbridge
   ```

#### Service-Verwaltung

```bash
# Status prüfen
sudo systemctl status modbridge

# Service starten
sudo systemctl start modbridge

# Service stoppen
sudo systemctl stop modbridge

# Service neustarten
sudo systemctl restart modbridge

# Logs ansehen (live)
sudo journalctl -u modbridge -f

# Logs ansehen (letzte 100 Zeilen)
sudo journalctl -u modbridge -n 100

# Service deaktivieren
sudo systemctl disable modbridge
```

#### Service-Konfiguration anpassen

Bearbeiten Sie `/etc/systemd/system/modbridge.service`:

```bash
sudo nano /etc/systemd/system/modbridge.service
```

Nach Änderungen:
```bash
sudo systemctl daemon-reload
sudo systemctl restart modbridge
```

---

### Methode 5: Mit Makefile

Für Entwickler, die häufig kompilieren und testen.

#### Verfügbare Make-Befehle

```bash
# Hilfe anzeigen (alle verfügbaren Befehle)
make help

# Kompilieren
make build

# Kompilieren und ausführen
make run

# Tests ausführen
make test

# Code formatieren
make fmt

# Linter ausführen
make lint

# Coverage-Report erstellen
make coverage

# Aufräumen (Build-Artefakte löschen)
make clean

# Docker Image bauen
make docker-build

# Docker Container starten
make docker-run

# Docker Container stoppen
make docker-stop

# Docker Logs ansehen
make docker-logs

# Abhängigkeiten aktualisieren
make deps

# Für alle Plattformen kompilieren
make build-all
```

#### Entwicklungs-Workflow

```bash
# Code bearbeiten, dann:
make fmt        # Code formatieren
make lint       # Code prüfen
make test       # Tests ausführen
make run        # Lokal testen
```

---

## Konfiguration

### Konfigurations-Datei (config.json)

Die Konfiguration wird in `config.json` im Arbeitsverzeichnis gespeichert.

**Speicherort je nach Installationsmethode**:
- **Docker**: `/app/config.json` (gemountet vom Host)
- **Systemd**: `/var/lib/modbridge/config.json`
- **Manuell**: `./config.json` (im aktuellen Verzeichnis)

#### Standard-Konfiguration

```json
{
  "web_port": ":8080",
  "admin_pass_hash": "",
  "proxies": []
}
```

#### Beispiel-Konfiguration

```json
{
  "web_port": ":8080",
  "admin_pass_hash": "$2a$14$GIEWZELPUk/ixrj.sEb12OTEJCgb6hpxWqA0mAzZVlbxYL5qZclnu",
  "proxies": [
    {
      "id": "21e71152-3866-43ac-891d-c5ec85fa1e98",
      "name": "Solar-Wechselrichter",
      "listen_addr": ":5020",
      "target_addr": "192.168.1.100:502",
      "enabled": true
    },
    {
      "id": "a7b8c9d0-1234-5678-90ab-cdef12345678",
      "name": "Energiezähler",
      "listen_addr": ":5021",
      "target_addr": "192.168.1.101:502",
      "enabled": true
    },
    {
      "id": "f1e2d3c4-b5a6-9786-5432-1fedcba09876",
      "name": "Batterie-Management",
      "listen_addr": ":5022",
      "target_addr": "192.168.1.102:502",
      "enabled": false
    }
  ]
}
```

### Konfigurations-Parameter erklärt

#### `web_port` (String)
- **Beschreibung**: Port für die Web-Oberfläche
- **Format**: `:PORT` oder `IP:PORT`
- **Standard**: `:8080`
- **Beispiele**:
  - `:8080` - Lauscht auf allen Interfaces, Port 8080
  - `127.0.0.1:8080` - Nur localhost
  - `:80` - Standard HTTP-Port (erfordert Root-Rechte)
  - `192.168.1.50:8080` - Spezifische IP-Adresse

#### `admin_pass_hash` (String)
- **Beschreibung**: Bcrypt-Hash des Admin-Passworts
- **Standard**: `""` (leer = Erstinstallation)
- **Hinweise**:
  - Wird automatisch beim ersten Login gesetzt
  - Bcrypt-Hashing mit Cost 14
  - **Niemals** im Klartext speichern
  - Kann über Web-UI geändert werden

#### `proxies` (Array)
Liste aller Proxy-Konfigurationen.

##### Proxy-Objekt-Parameter

###### `id` (String, UUID)
- **Beschreibung**: Eindeutige Proxy-ID
- **Format**: UUID v4
- **Beispiel**: `"21e71152-3866-43ac-891d-c5ec85fa1e98"`
- **Hinweise**: Wird automatisch generiert beim Erstellen

###### `name` (String)
- **Beschreibung**: Anzeigename des Proxies
- **Beispiele**:
  - `"Solar-Wechselrichter"`
  - `"Energiezähler Haupteingang"`
  - `"PV-Anlage 1"`

###### `listen_addr` (String)
- **Beschreibung**: Adresse, auf der der Proxy lauscht
- **Format**: `:PORT` oder `IP:PORT`
- **Beispiele**:
  - `:5020` - Lauscht auf allen Interfaces
  - `127.0.0.1:5020` - Nur localhost
  - `192.168.1.50:5020` - Spezifische IP
- **Hinweise**:
  - Muss eindeutig sein (kein Port doppelt)
  - Ports < 1024 erfordern Root-Rechte
  - Standard Modbus-Port: 502

###### `target_addr` (String)
- **Beschreibung**: Ziel-Adresse des Modbus-Geräts
- **Format**: `IP:PORT` oder `HOSTNAME:PORT`
- **Beispiele**:
  - `192.168.1.100:502`
  - `modbus-device.local:502`
  - `10.0.0.50:5020`
- **Hinweise**:
  - Muss erreichbar sein
  - DNS-Auflösung wird unterstützt
  - Standard Modbus-Port: 502

###### `enabled` (Boolean)
- **Beschreibung**: Ob der Proxy beim Start aktiviert wird
- **Werte**: `true` oder `false`
- **Standard**: `false`
- **Hinweise**:
  - Kann über Web-UI geändert werden
  - Deaktivierte Proxies werden gespeichert, aber nicht gestartet

### Konfiguration über Web-Interface

Die meisten Einstellungen können über die Web-Oberfläche verwaltet werden:

1. **Öffnen**: `http://localhost:8080`
2. **Login**: Mit Admin-Passwort (beim ersten Start festlegen)
3. **Navigation**:
   - **Dashboard**: Übersicht aller Proxies
   - **Proxies**: Proxies hinzufügen, bearbeiten, löschen
   - **Logs**: Echtzeit-Logging
   - **Configuration**: Export/Import von Konfigurationen

#### Proxy hinzufügen

1. Klicken Sie auf **"+ Add Proxy"**
2. Füllen Sie die Felder aus:
   - **Name**: Beschreibender Name
   - **Listen Address**: z.B. `:5020`
   - **Target Address**: z.B. `192.168.1.100:502`
3. Klicken Sie auf **"Save"**
4. Starten Sie den Proxy mit dem **Start**-Button

#### Konfiguration exportieren

1. Navigieren Sie zu **Configuration**
2. Klicken Sie auf **"Export"**
3. Die Datei `config.json` wird heruntergeladen

#### Konfiguration importieren

1. Navigieren Sie zu **Configuration**
2. Klicken Sie auf **"Import"**
3. Wählen Sie die `config.json` Datei
4. Bestätigen Sie den Import
5. **Hinweis**: Überschreibt aktuelle Einstellungen (außer Admin-Passwort)

---

## Verwendung

### Erste Schritte

1. **Anwendung starten** (je nach Installationsmethode):
   ```bash
   # Manuell
   ./modbridge

   # Docker
   docker-compose up -d

   # Systemd
   sudo systemctl start modbridge
   ```

2. **Web-Interface öffnen**:
   ```
   http://localhost:8080
   ```

3. **Beim ersten Start**:
   - Sie werden aufgefordert, ein Admin-Passwort zu setzen
   - Merken Sie sich dieses Passwort gut
   - Das Passwort wird mit Bcrypt gehashed gespeichert

4. **Login**:
   - Geben Sie Ihr Admin-Passwort ein
   - Session bleibt für 24h aktiv

### Dashboard

Das Dashboard zeigt:
- **Proxy-Status**: Anzahl laufender/gestoppter Proxies
- **Gesamtstatistiken**: Requests, Fehler, Uptime
- **Fehlerrate**: Prozentsatz fehlgeschlagener Anfragen
- **System-Informationen**: Version, Uptime

### Proxy-Verwaltung

#### Proxy erstellen

1. Klicken Sie auf **"Proxies"** in der Navigation
2. Klicken Sie auf **"+ Add Proxy"**
3. Füllen Sie das Formular aus:
   ```
   Name: Solar-Wechselrichter
   Listen Address: :5020
   Target Address: 192.168.1.100:502
   ```
4. Klicken Sie auf **"Save"**

#### Proxy starten

1. Finden Sie den Proxy in der Liste
2. Klicken Sie auf **"Start"**
3. Status ändert sich zu **"Running"**

#### Proxy stoppen

1. Finden Sie den laufenden Proxy
2. Klicken Sie auf **"Stop"**
3. Status ändert sich zu **"Stopped"**

#### Proxy löschen

1. Stoppen Sie den Proxy zuerst
2. Klicken Sie auf **"Delete"**
3. Bestätigen Sie die Löschung

### Logging

#### Live-Logs ansehen

1. Navigieren Sie zu **"Logs"**
2. Logs werden in Echtzeit angezeigt
3. Auto-Scroll ist standardmäßig aktiviert

#### Logs filtern

1. Verwenden Sie das **Suchfeld** für Textsuche
2. Verwenden Sie den **Level-Filter**:
   - **INFO**: Normale Betriebsmeldungen
   - **WARN**: Warnungen
   - **ERROR**: Fehler

#### Logs exportieren

1. Klicken Sie auf **"Download Logs"**
2. Die letzten 1000 Log-Einträge werden als JSON heruntergeladen

### Modbus-Client-Verbindung

Um sich mit einem Proxy zu verbinden:

```python
# Beispiel mit pymodbus
from pymodbus.client import ModbusTcpClient

# Verbindung zum Proxy (nicht direkt zum Gerät!)
client = ModbusTcpClient('localhost', port=5020)

# Normaler Modbus-Verkehr
result = client.read_holding_registers(0, 10, slave=1)
print(result.registers)

client.close()
```

**Wichtig**:
- Modbus-Clients verbinden sich mit dem **Proxy-Port** (z.B. 5020)
- Der Proxy leitet an das **Zielgerät** weiter (z.B. 192.168.1.100:502)
- Alle Modbus-Funktionen werden unterstützt

---

## Umgebungsvariablen

### Verfügbare Variablen

#### `WEB_PORT`
- **Beschreibung**: Überschreibt den Web-Port aus config.json
- **Format**: `:PORT` oder `IP:PORT`
- **Standard**: `:8080`
- **Beispiel**:
  ```bash
  export WEB_PORT=:9090
  ./modbridge
  ```

#### `LOG_LEVEL` (geplant für v0.2.0)
- **Beschreibung**: Log-Level festlegen
- **Werte**: `DEBUG`, `INFO`, `WARN`, `ERROR`
- **Standard**: `INFO`

#### `CONFIG_PATH` (geplant für v0.2.0)
- **Beschreibung**: Pfad zur Konfigurations-Datei
- **Standard**: `./config.json`

### Docker-Umgebungsvariablen

In `docker-compose.yml`:

```yaml
services:
  modbridge:
    environment:
      - WEB_PORT=:8080
      - LOG_LEVEL=INFO
```

Oder beim Docker-Run:

```bash
docker run -e WEB_PORT=:9090 -e LOG_LEVEL=DEBUG modbridge
```

### Systemd-Umgebungsvariablen

In `/etc/systemd/system/modbridge.service`:

```ini
[Service]
Environment="WEB_PORT=:8080"
Environment="LOG_LEVEL=INFO"
```

Nach Änderungen:
```bash
sudo systemctl daemon-reload
sudo systemctl restart modbridge
```

---

## Performance

### Benchmarks (v0.1.0)

**Test-System**: Intel i7-10700K, 16GB RAM, Ubuntu 22.04

| Metrik | Wert |
|--------|------|
| Latenz (avg) | ~3-5ms |
| Latenz (p95) | ~8ms |
| Latenz (p99) | ~12ms |
| Durchsatz | ~10,000 req/s |
| Speicher (idle) | ~8MB |
| Speicher (load) | ~15MB |
| CPU (idle) | <1% |
| CPU (load) | ~15-20% |
| Max. Connections | ~1,000 |

### Performance-Tuning

#### Linux System-Tuning

```bash
# File Descriptor Limits erhöhen
ulimit -n 65535

# TCP-Parameter optimieren
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
sudo sysctl -w net.ipv4.tcp_fin_timeout=30
sudo sysctl -w net.core.somaxconn=1024

# Permanent machen
sudo nano /etc/sysctl.conf
# Fügen Sie hinzu:
net.ipv4.tcp_tw_reuse=1
net.ipv4.tcp_fin_timeout=30
net.core.somaxconn=1024
```

#### Docker Performance

In `docker-compose.yml`:

```yaml
services:
  modbridge:
    ulimits:
      nofile:
        soft: 65535
        hard: 65535
```

#### Go Runtime-Tuning

```bash
# GOMAXPROCS auf CPU-Anzahl setzen
export GOMAXPROCS=4

# Garbage Collection tunen
export GOGC=100
```

### Kapazitätsplanung

**Für 1,000 gleichzeitige Verbindungen:**
- CPU: 2 Cores (min)
- RAM: 512MB (min)
- Netzwerk: 100Mbps

**Für 10,000 gleichzeitige Verbindungen:**
- CPU: 4-8 Cores
- RAM: 2GB
- Netzwerk: 1Gbps

**Für 100,000+ gleichzeitige Verbindungen:**
- CPU: 16+ Cores
- RAM: 8-16GB
- Netzwerk: 10Gbps
- Mehrere Instanzen empfohlen

### Performance-Monitoring

Detaillierte Performance-Anleitungen finden Sie in [docs/PERFORMANCE.md](docs/PERFORMANCE.md).

---

## Troubleshooting

### Häufige Probleme

#### Port bereits in Verwendung

**Problem**:
```
bind: address already in use
```

**Lösung**:
```bash
# Prozess finden, der den Port verwendet
sudo lsof -i :8080

# Prozess beenden
sudo kill -9 <PID>

# Oder anderen Port verwenden
export WEB_PORT=:9090
```

#### Keine Verbindung zum Zielgerät

**Problem**: Proxy kann Modbus-Gerät nicht erreichen

**Diagnose**:
```bash
# Ping-Test
ping 192.168.1.100

# Port-Test
nc -zv 192.168.1.100 502

# Traceroute
traceroute 192.168.1.100
```

**Lösungen**:
- Prüfen Sie Firewall-Regeln
- Prüfen Sie Netzwerk-Routing
- Prüfen Sie, ob Gerät erreichbar ist
- Prüfen Sie Modbus-Port (Standard: 502)

#### Permission Denied (Port < 1024)

**Problem**:
```
permission denied
```

**Lösung**:
```bash
# Entweder als Root ausführen
sudo ./modbridge

# Oder CAP_NET_BIND_SERVICE setzen
sudo setcap 'cap_net_bind_service=+ep' ./modbridge
./modbridge

# Oder Port >= 1024 verwenden
# In config.json: "web_port": ":8080"
```

#### Docker Container startet nicht

**Problem**: Container stoppt sofort nach Start

**Diagnose**:
```bash
# Logs ansehen
docker logs modbridge

# Container-Status prüfen
docker ps -a

# Detaillierte Inspection
docker inspect modbridge
```

**Lösungen**:
- Prüfen Sie config.json Syntax
- Prüfen Sie Port-Verfügbarkeit
- Prüfen Sie Volume-Mounts

#### Systemd Service startet nicht

**Problem**: Service failed to start

**Diagnose**:
```bash
# Status prüfen
sudo systemctl status modbridge

# Detaillierte Logs
sudo journalctl -u modbridge -n 100 --no-pager

# Service-Konfiguration prüfen
sudo systemctl cat modbridge
```

**Lösungen**:
- Prüfen Sie Binary-Pfad: `/opt/modbridge/modbridge`
- Prüfen Sie Berechtigungen
- Prüfen Sie Konfiguration
- Prüfen Sie systemd-Unit-Datei

### Log-Analyse

#### Log-Formate

**INFO-Logs**:
```
[2025-12-24T21:16:33Z] [INFO] SYSTEM: Starting Modbus Manager on :8080
[2025-12-24T21:16:35Z] [INFO] proxy-id: Started proxy listening on :5020 -> 192.168.1.100:502
```

**ERROR-Logs**:
```
[2025-12-24T21:20:15Z] [ERROR] proxy-id: Forward error: dial tcp 192.168.1.100:502: connect: connection refused
[2025-12-24T21:20:16Z] [ERROR] SYSTEM: Failed to start proxy: address already in use
```

#### Debugging aktivieren

(Geplant für v0.2.0)

```bash
export LOG_LEVEL=DEBUG
./modbridge
```

### Support

Bei Problemen:

1. **Logs prüfen**:
   - Web-UI: Logs-Seite
   - Docker: `docker logs modbridge`
   - Systemd: `journalctl -u modbridge -f`

2. **Konfiguration prüfen**:
   ```bash
   cat config.json | jq .
   ```

3. **Issue erstellen**:
   - GitHub: https://github.com/Xerolux/modbridge/issues
   - Fügen Sie hinzu:
     - Logs
     - Konfiguration (ohne Passwort-Hash!)
     - System-Informationen
     - Reproduktionsschritte

---

## Entwicklung

### Lokales Development

```bash
# Repository klonen
git clone https://github.com/Xerolux/modbridge.git
cd modbridge

# Abhängigkeiten laden
go mod download

# Code formatieren
make fmt

# Linter ausführen
make lint

# Tests ausführen
make test

# Lokal ausführen
make run
```

### Tests

```bash
# Alle Tests
go test ./...

# Mit Coverage
go test -cover ./...

# Detaillierter Coverage-Report
make coverage
```

### Beitragen

Bitte lesen Sie [CONTRIBUTING.md](CONTRIBUTING.md) für Details.

---

## Roadmap

Details zur zukünftigen Entwicklung finden Sie in [ROADMAP.md](ROADMAP.md).

### Geplante Features

- **v0.2.0**: Performance-Optimierungen, Connection Pooling
- **v0.3.0**: Prometheus Metrics, Grafana-Dashboard
- **v0.4.0**: SSL/TLS-Unterstützung, Multi-User-Support
- **v1.0.0**: Production-Ready Release

---

## Lizenz

MIT License - siehe [LICENSE](LICENSE) für Details.

---

## Autoren

- **Xerolux** - Initial work - [GitHub](https://github.com/Xerolux)

---

## Danksagungen

- Modbus-Protokoll-Implementierung basierend auf Modbus TCP/IP Spezifikation
- Web-UI inspiriert von modernen Design-Prinzipien
- Go-Community für exzellente Tools und Bibliotheken

---

**Version**: 0.1.0
**Letzte Aktualisierung**: 24. Dezember 2025
**Status**: Beta
