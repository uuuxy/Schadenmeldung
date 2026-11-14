#!/usr/bin/env bash
# ==============================================================================
# update.sh — holt den neuesten Stand von GitHub und spielt ihn ein, in einem Rutsch.
# Auf dem Server im Projektordner aufrufen: ./update.sh
# ==============================================================================
# Ablauf (wie update.sh der Bibliothek, schlanker):
#   1. Sicherung der Datenbank nach ./backups/ (nur für root lesbar)
#   2. Code holen: der Stand von origin/main
#   3. Bauen und starten: Basis-Images frisch, Datenbank in der neuesten Nebenversion
#   4. Warten, bis die App gesund antwortet
#   5. Aufräumen: Sicherungen älter als 7 Tage, alte Images dieser App
#   6. Zeitplan: nächtliche Sicherung und Wächter, falls sie noch fehlen (neuer Server)
# Geht etwas schief, bricht das Skript ab und nennt den Weg zurück.
# ==============================================================================
set -euo pipefail

# Alles steht in einer Klammer: Schritt 2 überschreibt dieses Skript, während es läuft. Bash liest
# eine Klammer ganz, bevor es sie ausführt; ohne sie liefe der Rest aus der neuen Datei weiter.
{

cd "$(dirname "$0")"
SICHERUNGEN="backups"
AUFBEWAHRUNG_TAGE=7
ZEIT="$(date +%Y%m%d_%H%M%S)"
SICHERUNG="${SICHERUNGEN}/vor_update_${ZEIT}.sql.gz"
VORHER=""
mkdir -p "${SICHERUNGEN}" && chmod 700 "${SICHERUNGEN}"

gruen='\033[0;32m'; gelb='\033[1;33m'; rot='\033[0;31m'; fett='\033[1m'; aus='\033[0m'
schritt() { echo -e "\n${fett}▸ $*${aus}"; }
gut()     { echo -e "${gruen}✓${aus} $*"; }
warnung() { echo -e "${gelb}!${aus} $*"; }
fehler()  { echo -e "${rot}✗ $*${aus}" >&2; }

rueckweg() {
    echo ""
    fehler "Update abgebrochen. Der Weg zurück:"
    if [ -n "${VORHER}" ]; then
        echo "  Alter Code:   git reset --hard ${VORHER} && docker compose up -d --build"
    fi
    if [ -f "${SICHERUNG}" ]; then
        echo "  Alte Daten (nur nötig, wenn das neue Programm die Datenbank schon umgebaut hat):"
        echo "    docker compose stop schaden"
        echo "    docker compose exec -T db dropdb -U schaden --force schaden"
        echo "    docker compose exec -T db createdb -U schaden schaden"
        echo "    gunzip -c ${SICHERUNG} | docker compose exec -T db psql -v ON_ERROR_STOP=1 -U schaden -d schaden"
        echo "    docker compose up -d"
    fi
    echo "  Protokoll:    journalctl -t schadenmeldung -n 50"
}

# Das Protokoll der App geht ins Journal des Servers (docker-compose.yml). Ohne journald startete der
# Container nicht; dann lieber gleich abbrechen, solange noch nichts geändert ist.
if ! systemctl is-active --quiet systemd-journald 2>/dev/null; then
    fehler "systemd-journald läuft nicht, die App könnte kein Protokoll schreiben. In docker-compose.yml beim Dienst schaden driver: json-file eintragen."
    exit 1
fi

# ── 1. Sicherung ──────────────────────────────────────────────────────────────
schritt "1. Datenbank sichern"
if docker compose ps --status running --services 2>/dev/null | grep -qx db; then
    # umask statt chmod danach: Sonst wäre die Datei lesbar, während sie geschrieben wird.
    if (umask 077; docker compose exec -T db pg_dump -U schaden -d schaden | gzip > "${SICHERUNG}"); then
        gut "Gesichert: ${SICHERUNG} ($(du -h "${SICHERUNG}" | cut -f1))"
    else
        rm -f "${SICHERUNG}"
        fehler "Die Sicherung ist fehlgeschlagen; es wurde nichts geändert."
        exit 1
    fi
else
    warnung "Die Datenbank läuft nicht (erster Start?) – ohne Sicherung weiter."
fi

# ── 2. Code holen ─────────────────────────────────────────────────────────────
schritt "2. Neuesten Stand von GitHub holen"
# Wie git pull, nur ohne Zusammenführen: Der Server bekommt genau origin/main, auch nachdem der
# Verlauf auf GitHub neu geschrieben wurde. Eigene Änderungen an Dateien des Repos gingen dabei
# verloren, deshalb bricht das Skript dann vorher ab. .env und backups/ sind nicht im Repo.
if ! git diff --quiet HEAD --; then
    fehler "In diesem Ordner sind Dateien des Repos verändert:"
    git status --short --untracked-files=no
    echo "Erst klären (git diff), dann erneut ./update.sh."
    exit 1
fi
VORHER="$(git rev-parse --short HEAD)"
git fetch --quiet origin main
git reset --quiet --hard origin/main
JETZT="$(git rev-parse --short HEAD)"
if [ "${VORHER}" = "${JETZT}" ]; then
    gut "Schon aktuell (${JETZT}) – gebaut wird trotzdem, für frische Basis-Images."
else
    gut "${VORHER} → ${JETZT}: $(git log -1 --format=%s)"
fi

# Schlüssel für das Passwort des Versand-Postfachs (Einstellungen → Mail): einmal zufällig erzeugt,
# bleibt er in der .env und damit außerhalb von Datenbank und Sicherungen.
if [ -f .env ] && ! grep -q '^SCHLUESSEL=.' .env; then
    sed -i '/^SCHLUESSEL=$/d' .env
    printf '\nSCHLUESSEL=%s\n' "$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')" >> .env
    gut "SCHLUESSEL in der .env angelegt."
fi

# ── 3. Bauen und starten ──────────────────────────────────────────────────────
schritt "3. Bauen und starten"
if docker compose pull --quiet db && docker compose build --pull schaden && docker compose up -d; then
    gut "Container laufen."
else
    rueckweg
    exit 1
fi

# ── 4. Gesundheit ─────────────────────────────────────────────────────────────
schritt "4. Warten, bis die App antwortet"
# Dieselbe Probe wie der Healthcheck im Dockerfile, aber sofort statt erst nach dessen 30 Sekunden.
GESUND=nein
for _ in $(seq 1 45); do
    if docker compose exec -T schaden /schaden gesund >/dev/null 2>&1; then
        GESUND=ja
        break
    fi
    sleep 2
done
if [ "${GESUND}" = ja ]; then
    gut "Die App antwortet (Stand ${JETZT})."
else
    fehler "Die App antwortet nach 90 Sekunden nicht."
    docker compose logs --tail 20 schaden 2>&1 | sed 's/^/    /'
    rueckweg
    exit 1
fi

# ── 5. Aufräumen ──────────────────────────────────────────────────────────────
schritt "5. Aufräumen"
# Die Sicherungen enthalten Namen und Fotos; länger als nötig liegen sie nicht (Datenschutzseite).
WEG=$(find "${SICHERUNGEN}" -name 'vor_update_*.sql.gz' -mtime "+${AUFBEWAHRUNG_TAGE}" -print -delete 2>/dev/null | wc -l | tr -d ' ')
gut "${WEG} Sicherung(en) älter als ${AUFBEWAHRUNG_TAGE} Tage gelöscht."
# Nur namenlose, unbenutzte Images (Reste früherer Bauten); benannte Images und alle Volumes
# bleiben, auch die der anderen Dienste auf diesem Server.
docker image prune -f >/dev/null 2>&1 || true

# ── 6. Zeitplan ───────────────────────────────────────────────────────────────
schritt "6. Zeitplan prüfen"
# Nächtliche Sicherung und Wächter laufen über cron. Fehlen sie, etwa auf einem neuen Server, legt das
# Update sie an, damit sie nicht vergessen werden; vorhandene Einträge bleiben, wie sie sind.
if command -v crontab >/dev/null; then
    HIER="$(pwd)"
    ZEITPLAN="$(crontab -l 2>/dev/null || true)"
    NEU="${ZEITPLAN}"
    grep -qF "${HIER}/sicherung.sh" <<<"${ZEITPLAN}" || NEU+=$'\n'"30 2 * * * ${HIER}/sicherung.sh 2>&1 | logger -t schadenmeldung"
    grep -qF "${HIER}/waechter.sh" <<<"${ZEITPLAN}" || NEU+=$'\n'"*/5 * * * * ${HIER}/waechter.sh 2>&1 | logger -t schadenmeldung"
    if [ "${NEU}" != "${ZEITPLAN}" ]; then
        sed '/^$/d' <<<"${NEU}" | crontab -
        gut "Zeitplan angelegt: Sicherung jede Nacht um 2:30, Wächter alle fünf Minuten."
    else
        gut "Zeitplan steht: nächtliche Sicherung und Wächter."
    fi
else
    warnung "cron fehlt: Sicherung und Wächter bitte von Hand einplanen (README)."
fi
echo -e "\n${gruen}${fett}Fertig.${aus} Speicher frei: $(df -h / | awk 'NR==2 {print $4}')"

exit 0
}
