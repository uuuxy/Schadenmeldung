#!/usr/bin/env bash
# sicherung.sh — sichert die Datenbank nach ./backups/ und löscht Sicherungen nach sieben Tagen.
# Für jede Nacht einmal auf dem Server einrichten (crontab -e als root, Pfad anpassen):
#   30 2 * * * /pfad/zur/schadenmeldung/sicherung.sh 2>&1 | logger -t schadenmeldung
# Ob es geklappt hat: journalctl -t schadenmeldung. Zurückspielen wie im Abbruch-Hinweis von update.sh.
set -euo pipefail
cd "$(dirname "$0")"

SICHERUNGEN="backups"
DATEI="${SICHERUNGEN}/nacht_$(date +%Y%m%d_%H%M%S).sql.gz"
mkdir -p "${SICHERUNGEN}" && chmod 700 "${SICHERUNGEN}"

# umask statt chmod danach: Sonst wäre die Datei lesbar, während sie geschrieben wird.
if ! (umask 077; docker compose exec -T db pg_dump -U schaden -d schaden | gzip > "${DATEI}"); then
    rm -f "${DATEI}"
    echo "Sicherung fehlgeschlagen" >&2
    exit 1
fi
# Die Sicherungen enthalten Namen und Fotos; länger als sieben Tage liegen sie nicht (Datenschutzseite).
find "${SICHERUNGEN}" -name 'nacht_*.sql.gz' -mtime +7 -delete
echo "Gesichert: ${DATEI} ($(du -h "${DATEI}" | cut -f1))"
