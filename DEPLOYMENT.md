# Deployment Guide

## Übersicht

Dieses Dokument beschreibt die optimierten Docker- und CI/CD-Konfigurationen für ModBridge.

## Docker Setup

### Dockerfile Optimierungen

Das neue Dockerfile verwendet Best Practices für Sicherheit und Performance:

- **Multi-Stage Build**: Kleineres finales Image (~20MB)
- **Alpine Linux 3.19**: Minimale Angriffsfläche
- **Non-root User**: Läuft als `appuser` (UID 1000)
- **Health Check**: Automatische Gesundheitsprüfung auf `/api/health`
- **Optimierte Build Flags**: `-ldflags="-s -w"` und `-trimpath` für kleinere Binaries

### Docker Compose Verbesserungen

Die aktualisierte `docker-compose.yml` bietet:

- **Named Volumes**: Persistente Datenspeicherung
- **Resource Limits**: CPU und Memory Constraints
- **Security Options**: `no-new-privileges` aktiviert
- **Logging Configuration**: Automatische Log-Rotation
- **Health Checks**: Integrierte Gesundheitsprüfungen
- **Environment Variables**: Flexible Konfiguration über `.env`

#### Verwendung

```bash
# Starten
docker compose up -d

# Logs ansehen
docker compose logs -f

# Stoppen
docker compose down

# Volumes entfernen
docker compose down -v
```

#### Environment Variables

Erstellen Sie eine `.env` Datei im Projektverzeichnis:

```env
LOG_LEVEL=INFO
TZ=Europe/Berlin
VERSION=latest
```

## CI/CD Pipeline

### Unified Workflow

Die neue `main.yml` kombiniert alle CI/CD-Prozesse in einem einzigen Workflow:

#### Jobs

1. **Quality** - Code-Qualitätsprüfungen
   - `go fmt` Formatierung
   - `go vet` Statische Analyse
   - `golangci-lint` Linting

2. **Security** - Sicherheitsscans
   - Gosec - Go Security Scanner
   - Trivy - Vulnerability Scanner
   - CodeQL - Code-Analyse

3. **Test** - Tests ausführen
   - Matrix-Tests mit Go 1.23 und 1.24
   - Race Detection
   - Code Coverage Upload zu Codecov

4. **Build** - Multi-Platform Builds
   - Linux (amd64, arm64)
   - Windows (amd64)
   - macOS (amd64, arm64)

5. **Docker** - Container Build & Push
   - Multi-Architecture (amd64, arm64)
   - Push zu GitHub Container Registry
   - Image Security Scanning

6. **Release** - Automatische Releases
   - Wird bei Tags ausgelöst (`v*`)
   - Erstellt GitHub Release
   - Uploaded Binaries und .deb Pakete
   - Generiert Changelog

### Workflow Trigger

- **Push**: main, master, develop Branches
- **Pull Request**: Zu main, master, develop
- **Tags**: v* (für Releases)
- **Manual**: workflow_dispatch

### Gelöschte Workflows

Die folgenden redundanten Workflows wurden entfernt:
- `ci.yml` - Ersetzt durch main.yml
- `docker-publish.yml` - Integriert in main.yml
- `release.yml` - Integriert in main.yml
- `pr-check.yml` - Integriert in main.yml
- `security.yml` - Integriert in main.yml
- `labeler.yml` - Nicht mehr benötigt
- `stale.yml` - Nicht mehr benötigt

## Build & Test Lokal

### Voraussetzungen

- Go 1.24 oder höher
- Docker (optional)
- Docker Compose (optional)

### Lokales Bauen

```bash
# Binary bauen
go build -ldflags="-s -w" -trimpath -o modbridge ./main.go

# Tests ausführen
go test -v -race -coverprofile=coverage.txt ./...

# Coverage ansehen
go tool cover -html=coverage.txt
```

### Docker Lokal

```bash
# Image bauen
docker build -t modbridge:local .

# Container starten
docker run -d -p 8080:8080 -p 5020-5030:5020-5030 modbridge:local

# Mit docker-compose
docker compose up -d
```

## Performance

### Docker Image Größe

- Builder Stage: ~800MB (wird nicht im finalen Image verwendet)
- Final Image: ~20MB
- Komprimiert: ~8MB

### Resource Limits (docker-compose.yml)

- **CPU Limit**: 2 Cores
- **CPU Reservation**: 0.5 Cores
- **Memory Limit**: 512MB
- **Memory Reservation**: 128MB

Diese können in `docker-compose.yml` angepasst werden.

## Sicherheit

### Docker Security

1. **Non-root User**: Container läuft als `appuser` (UID 1000)
2. **No New Privileges**: Verhindert Privilege Escalation
3. **Minimal Base Image**: Alpine Linux 3.19
4. **Security Scanning**: Trivy und Gosec in CI/CD

### Best Practices

- Verwenden Sie immer Tagged Versions statt `latest` in Produktion
- Halten Sie Dependencies aktuell
- Überprüfen Sie regelmäßig Security Scan Ergebnisse
- Verwenden Sie Secrets für sensitive Daten (nicht in config.json committen)

## Troubleshooting

### Docker Build schlägt fehl

```bash
# Cache löschen
docker builder prune -a

# Neu bauen
docker build --no-cache -t modbridge:local .
```

### Port bereits in Verwendung

```bash
# Port in docker-compose.yml ändern
ports:
  - "9090:8080"  # Statt 8080:8080
```

### Healthcheck schlägt fehl

```bash
# Logs prüfen
docker logs modbridge

# Healthcheck Status prüfen
docker inspect modbridge | grep Health -A 10
```

## Migration von alten Workflows

Wenn Sie von den alten Workflows migrieren:

1. **Secrets**: Keine zusätzlichen Secrets erforderlich (verwendet `GITHUB_TOKEN`)
2. **Branches**: Workflow funktioniert auf main, master und develop
3. **Releases**: Erstellen Sie einen Tag mit `v*` Format (z.B. `v0.2.0`)

```bash
# Release erstellen
git tag -a v0.2.0 -m "Release version 0.2.0"
git push origin v0.2.0
```

## Support

Bei Problemen:

1. Prüfen Sie die [GitHub Actions Logs](https://github.com/Xerolux/modbridge/actions)
2. Erstellen Sie ein [Issue](https://github.com/Xerolux/modbridge/issues)
3. Konsultieren Sie die [Dokumentation](README.md)
