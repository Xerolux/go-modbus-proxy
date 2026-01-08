# Release Checklist für v0.1.0

## Status: Bereit für Release! ✅

Alle Änderungen sind committed und gepusht auf Branch `claude/setup-docker-debian-uygKP`.

---

## Option 1: Pull Request erstellen und mergen (Empfohlen)

### Schritt 1: Pull Request erstellen

Öffne diese URL in deinem Browser:
```
https://github.com/Xerolux/modbridge/compare/main...claude/setup-docker-debian-uygKP?expand=1
```

### Schritt 2: PR Details ausfüllen

**Title**:
```
Add .deb packages and Docker automation
```

**Description**:
```markdown
## Summary
This PR adds comprehensive installation and deployment automation:

### New Features
✅ **Debian/Ubuntu Packages**
- .deb packages for AMD64 and ARM64
- Automatic systemd service setup
- Simple installation: `sudo dpkg -i modbridge_*.deb`

✅ **Docker Automation**
- Automated Docker builds on every push
- Multi-architecture support (AMD64 + ARM64)
- Public GitHub Container Registry (ghcr.io)
- Pre-built images ready to use

✅ **GitHub Actions Workflows**
- `release.yml`: Automated releases with .deb packages and Docker images
- `docker-publish.yml`: Continuous Docker builds
- Automatic version tagging and release notes

✅ **Documentation**
- `INSTALL_DEBIAN.md`: Comprehensive Debian/Ubuntu installation guide
- `.github/WORKFLOWS.md`: Complete CI/CD documentation
- Updated README with badges and quick install

### Installation Methods
1. **Debian/Ubuntu .deb** (easiest)
2. **Docker** (pre-built from ghcr.io)
3. **Docker Compose** (one command)
4. **From source** (for developers)

Ready to merge and create first release!
```

### Schritt 3: PR mergen

1. Klicke auf "Create Pull Request"
2. Warte auf CI-Checks (optional)
3. Klicke auf "Merge Pull Request"
4. Bestätige mit "Confirm Merge"

### Schritt 4: Release erstellen

Nach dem Merge in `main`:

```bash
# Checkout main branch
git checkout main
git pull origin main

# Tag erstellen
git tag -a v0.1.0 -m "Release v0.1.0: Initial release

Features:
- Multi-proxy Modbus TCP management
- Web-based UI with real-time monitoring
- Debian/Ubuntu .deb packages (AMD64 + ARM64)
- Docker images (Multi-arch: AMD64 + ARM64)
- systemd service integration
- Automated GitHub Actions workflows
"

# Tag pushen (triggert automatischen Release)
git push origin v0.1.0
```

### Schritt 5: Workflow beobachten

Öffne: https://github.com/Xerolux/modbridge/actions

Der Release-Workflow wird automatisch:
- ✅ Alle Binaries bauen (Linux, Windows, macOS)
- ✅ .deb Pakete bauen (AMD64 + ARM64)
- ✅ Docker Images bauen und pushen zu ghcr.io
- ✅ GitHub Release erstellen mit allen Dateien

**Dauer**: ~5-10 Minuten

### Schritt 6: Release veröffentlichen

Der Release ist automatisch verfügbar unter:
```
https://github.com/Xerolux/modbridge/releases/tag/v0.1.0
```

---

## Option 2: Direkter Merge und Release (Schneller)

Wenn du direkten Zugriff auf das Repository hast:

### Via GitHub Web Interface

1. **Branch mergen**:
   - Gehe zu: https://github.com/Xerolux/modbridge
   - Klicke auf "Compare & pull request" (falls sichtbar)
   - Oder erstelle PR über: https://github.com/Xerolux/modbridge/compare/main...claude/setup-docker-debian-uygKP

2. **Release erstellen**:
   - Gehe zu: https://github.com/Xerolux/modbridge/releases/new
   - Tag: `v0.1.0`
   - Release title: `Release v0.1.0`
   - Target: `main`
   - Description: (wird automatisch generiert)
   - Klicke "Publish release"

Die Workflows starten automatisch!

---

## Was passiert nach dem Release?

### Automatisch erstellt:

1. **GitHub Release** mit:
   - ✅ Binaries: `modbridge-linux-amd64`, `modbridge-linux-arm64`, etc.
   - ✅ .deb Pakete: `modbridge_0.1.0_amd64.deb`, `modbridge_0.1.0_arm64.deb`
   - ✅ Automatische Release Notes

2. **Docker Images** auf ghcr.io:
   - ✅ `ghcr.io/xerolux/modbridge:latest`
   - ✅ `ghcr.io/xerolux/modbridge:v0.1.0`
   - ✅ `ghcr.io/xerolux/modbridge:0.1`
   - ✅ `ghcr.io/xerolux/modbridge:0`

3. **Installation sofort verfügbar**:
   ```bash
   # .deb Package
   wget https://github.com/Xerolux/modbridge/releases/download/v0.1.0/modbridge_0.1.0_amd64.deb
   sudo dpkg -i modbridge_0.1.0_amd64.deb

   # Docker
   docker pull ghcr.io/xerolux/modbridge:latest
   docker run -d -p 8080:8080 ghcr.io/xerolux/modbridge:latest

   # Docker Compose
   docker-compose up -d  # verwendet automatisch ghcr.io
   ```

---

## Nächste Schritte nach Release

### Sofort testen:

1. **Docker Image testen**:
   ```bash
   docker pull ghcr.io/xerolux/modbridge:v0.1.0
   docker run -d -p 8080:8080 ghcr.io/xerolux/modbridge:v0.1.0
   # Öffne: http://localhost:8080
   ```

2. **.deb Package testen** (auf Debian/Ubuntu):
   ```bash
   wget https://github.com/Xerolux/modbridge/releases/download/v0.1.0/modbridge_0.1.0_amd64.deb
   sudo dpkg -i modbridge_0.1.0_amd64.deb
   sudo systemctl start modbridge
   sudo systemctl status modbridge
   # Öffne: http://localhost:8080
   ```

### Optional: Dokumentation erweitern

- README aktualisieren mit Release-Badge
- Screenshots hinzufügen
- Beispiel-Konfigurationen
- Video-Tutorial erstellen

---

## Troubleshooting

### Workflow schlägt fehl?

**Logs ansehen**:
https://github.com/Xerolux/modbridge/actions

**Häufige Probleme**:
- Permissions fehlen → Check Repository Settings → Actions → General
- Docker Push fehl → Check "packages: write" Permission
- .deb Build fehl → Check Makefile Syntax

### Tag löschen (falls nötig)

```bash
# Lokal
git tag -d v0.1.0

# Remote
git push --delete origin v0.1.0
```

---

## Alle Dateien bereit

✅ **Code**: Alle Änderungen committed
✅ **Workflows**: `.github/workflows/release.yml`, `docker-publish.yml`
✅ **Dokumentation**: `README.md`, `INSTALL_DEBIAN.md`, `.github/WORKFLOWS.md`
✅ **Build-System**: `Makefile` mit .deb-Targets
✅ **.deb Pakete**: In `releases/` (für Testing)
✅ **Docker**: `docker-compose.yml` nutzt ghcr.io

**Bereit für Release v0.1.0!** 🚀

---

**Branch**: `claude/setup-docker-debian-uygKP`
**Status**: Alle Tests bestanden ✅
**Nächster Schritt**: PR erstellen und mergen
