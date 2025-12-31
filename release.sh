#!/bin/bash

# ModBridge Release Script
# Automatisiert den gesamten Release-Prozess

set -e

# Farben für Output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Header
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   ModBridge Release Automatisierung   ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo ""

# Funktion: Banner anzeigen
banner() {
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  $1${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Funktion: Erfolg anzeigen
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Funktion: Warnung anzeigen
warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Funktion: Fehler anzeigen
error() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

# Funktion: Info anzeigen
info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Prüfe ob wir im richtigen Verzeichnis sind
if [ ! -f "version.txt" ]; then
    error "Nicht im modbridge-Verzeichnis! Bitte wechsle ins Projekt-Root."
fi

# Prüfe ob Git Repository
if [ ! -d ".git" ]; then
    error "Kein Git-Repository gefunden!"
fi

# Aktuelle Version auslesen
CURRENT_VERSION=$(cat version.txt 2>/dev/null || echo "0.0.0")
info "Aktuelle Version: ${CURRENT_VERSION}"
echo ""

# ═══════════════════════════════════════════════════════════
# Schritt 1: Versionsnummer eingeben
# ═══════════════════════════════════════════════════════════
banner "Schritt 1: Neue Version festlegen"

echo "Gib die neue Versionsnummer ein (z.B. 0.2.0, 1.0.0):"
read -p "Version: " NEW_VERSION

# Validiere Version (muss Format X.Y.Z haben)
if ! [[ $NEW_VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "Ungültiges Versionsformat! Verwende X.Y.Z (z.B. 0.2.0)"
fi

success "Neue Version: ${NEW_VERSION}"
echo ""

# ═══════════════════════════════════════════════════════════
# Schritt 2: Release-Typ auswählen
# ═══════════════════════════════════════════════════════════
banner "Schritt 2: Release-Typ auswählen"

echo "Wähle den Release-Typ:"
echo "  1) Stable Release (Production)"
echo "  2) Beta Release"
echo "  3) Release Candidate (RC)"
echo "  4) Draft (Entwurf)"
echo ""
read -p "Auswahl [1-4]: " RELEASE_TYPE_CHOICE

case $RELEASE_TYPE_CHOICE in
    1)
        RELEASE_TYPE="stable"
        TAG_SUFFIX=""
        PRERELEASE=false
        DRAFT=false
        success "Release-Typ: Stable (Production)"
        ;;
    2)
        RELEASE_TYPE="beta"
        TAG_SUFFIX="-beta"
        PRERELEASE=true
        DRAFT=false
        success "Release-Typ: Beta"
        ;;
    3)
        RELEASE_TYPE="rc"
        TAG_SUFFIX="-rc"
        PRERELEASE=true
        DRAFT=false
        success "Release-Typ: Release Candidate"
        ;;
    4)
        RELEASE_TYPE="draft"
        TAG_SUFFIX=""
        PRERELEASE=false
        DRAFT=true
        success "Release-Typ: Draft (Entwurf)"
        ;;
    *)
        error "Ungültige Auswahl!"
        ;;
esac

FULL_VERSION="${NEW_VERSION}${TAG_SUFFIX}"
TAG_NAME="v${FULL_VERSION}"

echo ""
info "Git Tag: ${TAG_NAME}"
echo ""

# ═══════════════════════════════════════════════════════════
# Schritt 3: Änderungen prüfen
# ═══════════════════════════════════════════════════════════
banner "Schritt 3: Git-Status prüfen"

# Prüfe auf uncommitted changes
if ! git diff-index --quiet HEAD --; then
    warning "Es gibt uncommitted Änderungen!"
    git status --short
    echo ""
    read -p "Möchtest du fortfahren? [j/N]: " CONTINUE
    if [[ ! $CONTINUE =~ ^[jJ]$ ]]; then
        error "Abgebrochen durch Benutzer"
    fi
fi

success "Git-Status geprüft"
echo ""

# ═══════════════════════════════════════════════════════════
# Schritt 4: Dateien aktualisieren
# ═══════════════════════════════════════════════════════════
banner "Schritt 4: Versions-Dateien aktualisieren"

# version.txt aktualisieren
echo "${NEW_VERSION}" > version.txt
success "version.txt → ${NEW_VERSION}"

# README.md aktualisieren
if [ -f "README.md" ]; then
    # Ersetze alte Version mit neuer in README
    sed -i "s/Version:** ${CURRENT_VERSION}/Version:** ${NEW_VERSION}/g" README.md
    sed -i "s/releases\/download\/v${CURRENT_VERSION}/releases\/download\/v${NEW_VERSION}/g" README.md
    sed -i "s/modbus-proxy-manager_${CURRENT_VERSION}/modbus-proxy-manager_${NEW_VERSION}/g" README.md
    success "README.md aktualisiert"
fi

# INSTALL_DEBIAN.md aktualisieren
if [ -f "INSTALL_DEBIAN.md" ]; then
    sed -i "s/releases\/download\/v${CURRENT_VERSION}/releases\/download\/v${NEW_VERSION}/g" INSTALL_DEBIAN.md
    sed -i "s/modbus-proxy-manager_${CURRENT_VERSION}/modbus-proxy-manager_${NEW_VERSION}/g" INSTALL_DEBIAN.md
    success "INSTALL_DEBIAN.md aktualisiert"
fi

# .github/WORKFLOWS.md aktualisieren
if [ -f ".github/WORKFLOWS.md" ]; then
    sed -i "s/Version: ${CURRENT_VERSION}/Version: ${NEW_VERSION}/g" .github/WORKFLOWS.md
    success ".github/WORKFLOWS.md aktualisiert"
fi

echo ""

# ═══════════════════════════════════════════════════════════
# Schritt 5: Zusammenfassung anzeigen
# ═══════════════════════════════════════════════════════════
banner "Schritt 5: Zusammenfassung"

echo "Release-Details:"
echo "  • Version:     ${NEW_VERSION}"
echo "  • Tag:         ${TAG_NAME}"
echo "  • Typ:         ${RELEASE_TYPE}"
echo "  • Prerelease:  ${PRERELEASE}"
echo "  • Draft:       ${DRAFT}"
echo ""

echo "Aktualisierte Dateien:"
git status --short
echo ""

read -p "Möchtest du fortfahren? [j/N]: " CONFIRM
if [[ ! $CONFIRM =~ ^[jJ]$ ]]; then
    error "Abgebrochen durch Benutzer"
fi

echo ""

# ═══════════════════════════════════════════════════════════
# Schritt 6: Git Commit & Tag erstellen
# ═══════════════════════════════════════════════════════════
banner "Schritt 6: Git Commit & Tag erstellen"

# Änderungen committen
git add version.txt README.md INSTALL_DEBIAN.md .github/WORKFLOWS.md
git commit -m "Bump version to ${NEW_VERSION}" || info "Nichts zu committen"
success "Änderungen committed"

# Git Tag erstellen
TAG_MESSAGE="Release ${TAG_NAME}"
if [ "$RELEASE_TYPE" == "beta" ]; then
    TAG_MESSAGE="Beta Release ${TAG_NAME}"
elif [ "$RELEASE_TYPE" == "rc" ]; then
    TAG_MESSAGE="Release Candidate ${TAG_NAME}"
elif [ "$RELEASE_TYPE" == "draft" ]; then
    TAG_MESSAGE="Draft Release ${TAG_NAME}"
fi

git tag -a "${TAG_NAME}" -m "${TAG_MESSAGE}"
success "Git Tag erstellt: ${TAG_NAME}"

echo ""

# ═══════════════════════════════════════════════════════════
# Schritt 7: Push zu GitHub
# ═══════════════════════════════════════════════════════════
banner "Schritt 7: Push zu GitHub"

echo "Möchtest du jetzt zu GitHub pushen?"
echo "  • Commit wird gepusht"
echo "  • Tag wird gepusht (startet automatisch Release-Workflow)"
echo ""
read -p "Jetzt pushen? [j/N]: " DO_PUSH

if [[ $DO_PUSH =~ ^[jJ]$ ]]; then
    # Aktuellen Branch ermitteln
    CURRENT_BRANCH=$(git branch --show-current)

    # Commit pushen
    info "Pushe zu Branch: ${CURRENT_BRANCH}"
    if git push origin "${CURRENT_BRANCH}"; then
        success "Commit gepusht"
    else
        error "Push fehlgeschlagen! Versuche manuell: git push origin ${CURRENT_BRANCH}"
    fi

    # Tag pushen (startet Release-Workflow!)
    info "Pushe Tag: ${TAG_NAME}"
    if git push origin "${TAG_NAME}"; then
        success "Tag gepusht - Release-Workflow startet jetzt!"
    else
        error "Tag-Push fehlgeschlagen! Versuche manuell: git push origin ${TAG_NAME}"
    fi
else
    warning "Push übersprungen. Führe manuell aus:"
    echo ""
    echo "  git push origin $(git branch --show-current)"
    echo "  git push origin ${TAG_NAME}"
fi

echo ""

# ═══════════════════════════════════════════════════════════
# Fertig!
# ═══════════════════════════════════════════════════════════
banner "Release ${TAG_NAME} erfolgreich erstellt! 🎉"

echo ""
echo "Nächste Schritte:"
echo ""

if [[ $DO_PUSH =~ ^[jJ]$ ]]; then
    echo "1. Workflow beobachten:"
    echo "   https://github.com/Xerolux/modbridge/actions"
    echo ""
    echo "2. Release wird automatisch erstellt in ~5-10 Minuten"
    echo ""
    echo "3. Verfügbar unter:"
    echo "   https://github.com/Xerolux/modbridge/releases/tag/${TAG_NAME}"
    echo ""
    echo "4. Docker Image wird verfügbar sein:"
    echo "   docker pull ghcr.io/xerolux/modbridge:${FULL_VERSION}"
    if [ "$RELEASE_TYPE" == "stable" ]; then
        echo "   docker pull ghcr.io/xerolux/modbridge:latest"
    fi
    echo ""
    echo "5. .deb Pakete werden erstellt:"
    echo "   modbus-proxy-manager_${NEW_VERSION}_amd64.deb"
    echo "   modbus-proxy-manager_${NEW_VERSION}_arm64.deb"
else
    echo "1. Änderungen und Tag pushen:"
    echo "   git push origin $(git branch --show-current)"
    echo "   git push origin ${TAG_NAME}"
    echo ""
    echo "2. Dann Workflow beobachten:"
    echo "   https://github.com/Xerolux/modbridge/actions"
fi

echo ""
success "Fertig!"
echo ""
