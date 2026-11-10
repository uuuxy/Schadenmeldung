#!/usr/bin/env bash
# ==============================================================================
# waechter.sh — schaut alle fünf Minuten nach der Schadensmeldung und sagt Bescheid, wenn etwas nicht
# stimmt. Einmal auf dem Server einrichten (crontab -e als root, Pfad anpassen):
#   */5 * * * * /pfad/zur/schadenmeldung/waechter.sh 2>&1 | logger -t schadenmeldung
# Ob die Mails ankommen, ohne echten Ausfall: ./waechter.sh --probe
# ==============================================================================
# Docker startet die App neu, wenn sie abstürzt, aber nicht, wenn sie hängt, und sagt niemandem
# Bescheid. Darum prüft dieses Skript:
#   - App und Datenbank: Sind sie beim zweiten Blick (nach 5 Minuten) noch gestört, startet es sie neu;
#     hilft das nicht, geht beim dritten Blick eine Mail raus und eine zweite, sobald alles wieder läuft.
#   - Die Platte: ab 90 % belegt eine Mail, denn voll kann die Datenbank nichts mehr speichern.
#   - Die nächtliche Sicherung (sicherung.sh): Ist die neueste älter als 26 Stunden, eine Mail.
# Jede Mail kommt einmal, bis der Grund behoben ist. Sie geht an WAECHTER_AN, sonst an den ersten aus
# ADMINS, über die Mail-Zugangsdaten der App aus .env, ohne weiteren Dienst.
# ==============================================================================
set -euo pipefail
cd "$(dirname "$0")"

STAND=".waechter"     # was schon gemeldet ist, von Lauf zu Lauf
PLATTE_GRENZE=90      # Prozent
SICHERUNG_STUNDEN=26  # eine Nacht und etwas Luft

# Ein Wert aus .env, ohne die Datei als Skript auszuführen.
aus_env() { sed -n "s/^$1=//p" .env 2>/dev/null | tail -n 1 | sed -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'$/\1/"; }

# melde BETREFF TEXT — verschlüsselt wie die App: Port 465 gleich über TLS, sonst STARTTLS als Pflicht,
# ohne nur lokal (Mailpit).
melde() {
    local an host port absender benutzer url betreff netrc="" schutz=(--ssl-reqd) zugang=()
    an="$(aus_env WAECHTER_AN)"
    [ -n "$an" ] || an="$(aus_env ADMINS | cut -d, -f1 | tr -d ' ')"
    host="$(aus_env SMTP_HOST)"
    port="$(aus_env SMTP_PORT)"
    port="${port:-587}"
    absender="$(aus_env MAIL_ABSENDER)"
    benutzer="$(aus_env SMTP_BENUTZER)"
    if [ -z "$an" ] || [ -z "$host" ] || [ -z "$absender" ]; then
        echo "Keine Mail möglich (WAECHTER_AN/ADMINS, SMTP_HOST oder MAIL_ABSENDER fehlt): $1" >&2
        return 0
    fi
    [ "$(aus_env APP_ENV)" = local ] && schutz=()
    if [ "$port" = 465 ]; then url="smtps://$host:$port"; else url="smtp://$host:$port"; fi
    if [ -n "$benutzer" ]; then
        # Das Passwort in einer Datei nur für root, nicht auf der Befehlszeile: Die sähe jeder in ps.
        netrc="$(mktemp)"
        printf 'machine %s login %s password %s\n' "$host" "$benutzer" "$(aus_env SMTP_PASSWORT)" > "$netrc"
        zugang=(--netrc-file "$netrc")
    fi
    betreff="=?UTF-8?B?$(printf '%s' "$1" | base64 | tr -d '\n')?="
    if ! printf 'From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n' \
            "$absender" "$an" "$betreff" "$2" |
        curl -sS --max-time 30 --url "$url" ${schutz[@]+"${schutz[@]}"} ${zugang[@]+"${zugang[@]}"} \
            --mail-from "$absender" --mail-rcpt "$an" --upload-file -; then
        echo "Mail an $an ging nicht raus: $1" >&2
    else
        echo "Mail an $an: $1"
    fi
    if [ -n "$netrc" ]; then rm -f "$netrc"; fi
}

if [ "${1:-}" = --probe ]; then
    melde "Schadensmeldung: Probe vom Wächter" "Diese Mail kommt vom Wächter (waechter.sh). Kommt sie an, kommt im Ernstfall auch die Warnung an."
    exit 0
fi

# zustand DIENST — etwa „running healthy“, „running starting“ oder „exited“; „fehlt“ ohne Container.
zustand() {
    local id
    id="$(docker compose ps -aq "$1" 2>/dev/null | head -n 1)"
    if [ -z "$id" ]; then echo fehlt; return; fi
    docker inspect -f '{{.State.Status}}{{if .State.Health}} {{.State.Health.Status}}{{end}}' "$id" 2>/dev/null || echo unbekannt
}

{ read -r blicke gemeldet platte sicherung < "$STAND"; } 2>/dev/null || { blicke=0; gemeldet=0; platte=0; sicherung=0; }

# ── App und Datenbank ─────────────────────────────────────────────────────────
kaputt=()
wie=""
for dienst in db schaden; do
    z="$(zustand "$dienst")"
    case "$z" in
        "running healthy" | "running starting" | running) ;;
        *) kaputt+=("$dienst"); wie="$wie $dienst: $z;" ;;
    esac
done
if [ "${#kaputt[@]}" -eq 0 ]; then
    if [ "$gemeldet" = 1 ]; then
        melde "Schadensmeldung läuft wieder" "Die Schadensmeldung läuft wieder normal (seit $(date '+%d.%m.%Y, %H:%M'))."
    fi
    blicke=0
    gemeldet=0
else
    blicke=$((blicke + 1))
    echo "Gestört (Blick ${blicke}):${wie}"
    if [ "$blicke" -eq 2 ]; then
        # Oft hängt nur die App; ein Neustart behebt das, bevor es jemand merkt.
        echo "Starte neu: ${kaputt[*]}"
        docker compose up -d "${kaputt[@]}" >/dev/null 2>&1 || true
        docker compose restart "${kaputt[@]}" >/dev/null 2>&1 || true
    elif [ "$blicke" -ge 3 ] && [ "$gemeldet" != 1 ]; then
        melde "Schadensmeldung gestört" "Die Schadensmeldung ist seit rund $(((blicke - 1) * 5)) Minuten gestört; ein Neustart hat nicht geholfen.
Betroffen:${wie}

Nachsehen auf dem Server im Projektordner:
  docker compose ps
  docker compose logs --tail 50 schaden"
        gemeldet=1
    fi
fi

# ── Platte ────────────────────────────────────────────────────────────────────
ordner="$(docker info -f '{{.DockerRootDir}}' 2>/dev/null || true)"
[ -d "$ordner" ] || ordner=/
belegt="$(df -P "$ordner" | awk 'NR == 2 { sub("%", "", $5); print $5 }')"
if [ "${belegt:-0}" -ge "$PLATTE_GRENZE" ]; then
    if [ "$platte" != 1 ]; then
        melde "Schadensmeldung: Platte fast voll" "Die Platte des Servers ist zu ${belegt} % belegt. Ist sie voll, kann die Datenbank nichts mehr speichern.
Platz schaffen, etwa mit: docker image prune"
        platte=1
    fi
else
    platte=0
fi

# ── Nächtliche Sicherung ──────────────────────────────────────────────────────
# Erst, wenn es überhaupt nächtliche Sicherungen gibt: Auf einem Testserver ohne sicherung.sh meldete
# der Wächter sonst jeden Tag.
if compgen -G "backups/nacht_*.sql.gz" >/dev/null; then
    if [ -z "$(find backups -name 'nacht_*.sql.gz' -mmin -$((SICHERUNG_STUNDEN * 60)) | head -n 1)" ]; then
        if [ "$sicherung" != 1 ]; then
            melde "Schadensmeldung: Sicherung fehlt" "Die neueste nächtliche Sicherung ist älter als ${SICHERUNG_STUNDEN} Stunden.
Nachsehen auf dem Server: journalctl -t schadenmeldung"
            sicherung=1
        fi
    else
        sicherung=0
    fi
fi

echo "$blicke $gemeldet $platte $sicherung" > "$STAND"
