# Release Automatisierung

## Automatisches Release-Script

Mit `release.sh` kannst du mit nur einem Befehl ein komplettes Release erstellen!

---

## 🚀 Quick Start

```bash
# Release-Script ausführen
./release.sh
```

Das Script führt dich interaktiv durch alle Schritte!

---

## 📋 Was das Script macht

### Automatisch:

1. ✅ **Versionsnummer eingeben** - Interaktive Eingabe (z.B. 0.2.0)
2. ✅ **Release-Typ wählen**:
   - Stable (Production)
   - Beta
   - Release Candidate (RC)
   - Draft (Entwurf)
3. ✅ **Alle Dateien aktualisieren**:
   - `version.txt`
   - `README.md`
   - `INSTALL_DEBIAN.md`
   - `.github/WORKFLOWS.md`
4. ✅ **Git Commit erstellen** - "Bump version to X.Y.Z"
5. ✅ **Git Tag erstellen** - z.B. `v0.2.0`, `v0.2.0-beta`
6. ✅ **Push zu GitHub** - Optional, mit Bestätigung
7. ✅ **Release-Workflow starten** - Automatisch beim Tag-Push

### GitHub Actions erstellt dann:

- Binaries für alle Plattformen
- .deb Pakete (AMD64 + ARM64)
- Docker Images (Multi-Arch)
- GitHub Release mit allen Dateien

---

## 🎯 Beispiel-Nutzung

### Beispiel 1: Stable Release

```bash
$ ./release.sh

╔════════════════════════════════════════╗
║   ModBridge Release Automatisierung   ║
╔════════════════════════════════════════╗

ℹ Aktuelle Version: 0.1.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Schritt 1: Neue Version festlegen
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Gib die neue Versionsnummer ein (z.B. 0.2.0, 1.0.0):
Version: 0.2.0
✓ Neue Version: 0.2.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Schritt 2: Release-Typ auswählen
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Wähle den Release-Typ:
  1) Stable Release (Production)
  2) Beta Release
  3) Release Candidate (RC)
  4) Draft (Entwurf)

Auswahl [1-4]: 1
✓ Release-Typ: Stable (Production)

ℹ Git Tag: v0.2.0

[... weitere Schritte ...]
```

### Beispiel 2: Beta Release

```bash
$ ./release.sh

[... Schritte 1-2 ...]

Auswahl [1-4]: 2
✓ Release-Typ: Beta

ℹ Git Tag: v0.2.0-beta
```

### Beispiel 3: Draft Release

```bash
$ ./release.sh

[... Schritte 1-2 ...]

Auswahl [1-4]: 4
✓ Release-Typ: Draft (Entwurf)

ℹ Git Tag: v0.2.0
```

---

## 🔧 Features im Detail

### 1. Versionsnummer-Validierung

Das Script akzeptiert nur gültige Semantic Versioning Nummern:

✅ Gültig:
- `0.2.0`
- `1.0.0`
- `2.5.3`

❌ Ungültig:
- `0.2` (fehlt Patch-Version)
- `v0.2.0` (kein 'v' Prefix)
- `1.0.0-beta` (Suffix wird automatisch hinzugefügt)

### 2. Release-Typen

#### Stable (Production)
- Git Tag: `v0.2.0`
- Prerelease: Nein
- Draft: Nein
- Docker Tags: `latest`, `0.2.0`, `0.2`, `0`

#### Beta
- Git Tag: `v0.2.0-beta`
- Prerelease: Ja
- Draft: Nein
- Docker Tags: `0.2.0-beta`

#### Release Candidate (RC)
- Git Tag: `v0.2.0-rc`
- Prerelease: Ja
- Draft: Nein
- Docker Tags: `0.2.0-rc`

#### Draft (Entwurf)
- Git Tag: `v0.2.0`
- Prerelease: Nein
- Draft: Ja (nur in GitHub sichtbar)
- Docker Tags: Keine (Draft wird nicht veröffentlicht)

### 3. Automatische Datei-Updates

Das Script aktualisiert automatisch alle relevanten Dateien:

**version.txt**:
```diff
- 0.1.0
+ 0.2.0
```

**README.md**:
```diff
- **Version:** 0.1.0
+ **Version:** 0.2.0

- wget .../releases/download/v0.1.0/modbus-proxy-manager_0.1.0_amd64.deb
+ wget .../releases/download/v0.2.0/modbus-proxy-manager_0.2.0_amd64.deb
```

**INSTALL_DEBIAN.md**:
```diff
- wget .../releases/download/v0.1.0/...
+ wget .../releases/download/v0.2.0/...
```

### 4. Git Integration

Automatisch:
- Commit mit Message: `Bump version to X.Y.Z`
- Annotated Tag mit Release-Message
- Optional: Push zu GitHub

### 5. Sicherheits-Checks

- ✅ Prüft ob im richtigen Verzeichnis
- ✅ Prüft auf Git-Repository
- ✅ Warnt bei uncommitted Changes
- ✅ Erfordert Bestätigung vor Push
- ✅ Validiert Versionsnummer

---

## 📊 Workflow nach Release

### Was passiert automatisch:

1. **Tag Push** → Startet Release Workflow
2. **GitHub Actions**:
   - Baut Binaries (5 Plattformen)
   - Baut .deb Pakete (2 Architekturen)
   - Baut Docker Images (Multi-Arch)
   - Erstellt GitHub Release
3. **Docker Registry**:
   - Pusht zu `ghcr.io/xerolux/modbridge`
   - Tags: `latest`, `v0.2.0`, `0.2`, `0`
4. **GitHub Release**:
   - Alle Artifacts verfügbar
   - Automatische Release Notes
   - Download-Links

**Dauer**: ~5-10 Minuten

---

## 🛠️ Troubleshooting

### Script schlägt fehl

**Problem**: `permission denied`

**Lösung**:
```bash
chmod +x release.sh
```

---

**Problem**: `Nicht im modbridge-Verzeichnis`

**Lösung**:
```bash
cd /pfad/zu/modbridge
./release.sh
```

---

**Problem**: `Ungültiges Versionsformat`

**Lösung**: Verwende Format `X.Y.Z` (z.B. `0.2.0`, nicht `0.2` oder `v0.2.0`)

---

### Git Push schlägt fehl

**Problem**: `error: RPC failed; HTTP 403`

**Lösung 1**: Branch-Name muss `claude/*` Format haben:
```bash
git checkout -b claude/release-v0.2.0
./release.sh
```

**Lösung 2**: Manueller Push:
```bash
git push origin main
git push origin v0.2.0
```

---

### Tag existiert bereits

**Problem**: `tag 'v0.2.0' already exists`

**Lösung 1**: Tag löschen und neu erstellen:
```bash
git tag -d v0.2.0
git push --delete origin v0.2.0
./release.sh
```

**Lösung 2**: Andere Version verwenden (z.B. `0.2.1`)

---

## 🔄 Release rückgängig machen

### Tag löschen (lokal)
```bash
git tag -d v0.2.0
```

### Tag löschen (remote)
```bash
git push --delete origin v0.2.0
```

### Release löschen (GitHub)
Gehe zu: `https://github.com/Xerolux/modbridge/releases`
- Klicke auf Release
- Klicke "Delete this release"

### Docker Image bleibt
Docker Images auf ghcr.io bleiben erhalten. Lösche sie manuell:
- Gehe zu GitHub → Packages
- Wähle Image-Version
- Delete

---

## 📝 Versionierung Best Practices

### Semantic Versioning (X.Y.Z)

- **X (Major)**: Breaking Changes
  - `0.1.0` → `1.0.0` - Erste Production Version
  - `1.2.0` → `2.0.0` - Breaking API Change

- **Y (Minor)**: Neue Features (backwards compatible)
  - `0.1.0` → `0.2.0` - Neue Features hinzugefügt
  - `1.0.0` → `1.1.0` - Feature-Update

- **Z (Patch)**: Bugfixes (backwards compatible)
  - `0.1.0` → `0.1.1` - Bugfix
  - `1.0.0` → `1.0.1` - Security Fix

### Release-Zyklus

```
Entwicklung → Beta → Release Candidate → Stable
    0.2.0-dev → 0.2.0-beta → 0.2.0-rc → 0.2.0
```

**Workflow**:
1. Entwickle Features auf `develop` Branch
2. Erstelle `0.2.0-beta` für Testing
3. Bei Erfolg: `0.2.0-rc` für Final Testing
4. Bei Erfolg: `0.2.0` Stable Release

---

## 🎯 Beispiel: Kompletter Release-Zyklus

### 1. Beta Release
```bash
./release.sh
# Version: 0.2.0
# Typ: 2 (Beta)
```

→ Tag: `v0.2.0-beta`
→ Testen, Bugs fixen

### 2. Release Candidate
```bash
./release.sh
# Version: 0.2.0
# Typ: 3 (RC)
```

→ Tag: `v0.2.0-rc`
→ Final Testing

### 3. Stable Release
```bash
./release.sh
# Version: 0.2.0
# Typ: 1 (Stable)
```

→ Tag: `v0.2.0`
→ Production Ready! 🎉

---

## 📚 Weitere Informationen

- **GitHub Actions**: `.github/WORKFLOWS.md`
- **Debian Installation**: `INSTALL_DEBIAN.md`
- **Docker Setup**: `README.md` → Docker Sektion
- **Makefile Targets**: `make help`

---

**Erstellt**: 31. Dezember 2025
**Version**: 1.0.0
