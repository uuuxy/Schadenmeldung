#!/usr/bin/env bash
# sicherung-pruefen.sh — spielt die neueste Sicherung (oder die angegebene Datei) probeweise in eine
# eigene Datenbank „probe“ ein, zählt Meldungen und Räume und löscht die Probe wieder. Die echte
# Datenbank bleibt unberührt. Einmal nach dem Einrichten und dann ab und zu, denn erst eine Sicherung,
# die sich zurückspielen ließ, ist eine.
#   ./sicherung-pruefen.sh                     # die neueste aus backups/
#   ./sicherung-pruefen.sh backups/nacht_….sql.gz
set -euo pipefail
cd "$(dirname "$0")"

# Die Dateinamen erzeugen sicherung.sh und update.sh selbst, ohne Leerzeichen.
# shellcheck disable=SC2012
datei="${1:-$(ls -1t backups/*.sql.gz 2>/dev/null | head -n 1)}"
if [ -z "$datei" ] || [ ! -f "$datei" ]; then
    echo "Keine Sicherung gefunden (backups/*.sql.gz)." >&2
    exit 1
fi

echo "Probe mit ${datei} …"
docker compose exec -T db dropdb -U schaden --if-exists probe
docker compose exec -T db createdb -U schaden probe
trap 'docker compose exec -T db dropdb -U schaden --if-exists probe >/dev/null 2>&1 || true' EXIT
gunzip -c "$datei" | docker compose exec -T db psql -q -v ON_ERROR_STOP=1 -U schaden -d probe >/dev/null
zahlen="$(docker compose exec -T db psql -U schaden -d probe -At \
    -c "SELECT (SELECT count(*) FROM meldungen) || ' Meldungen, ' || (SELECT count(*) FROM raeume) || ' Räume'")"
echo "✓ Die Sicherung lässt sich zurückspielen: ${zahlen}."
