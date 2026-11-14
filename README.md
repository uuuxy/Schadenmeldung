# Schadensmeldung

Schäden im Schulgebäude vom Handy melden, der Hausmeister bearbeitet sie. Ein Go-Programm mit
eingebetteter Svelte-Oberfläche und PostgreSQL; angemeldet wird mit dem Postfach der Schule (IMAP).
Lizenz: [EUPL-1.2](LICENSE).

## Betrieb

- `.env` aus `.env.example` anlegen, dann `./update.sh`: sichert die Datenbank, holt `origin/main`,
  baut, startet und prüft. Dasselbe bei jedem Update. Fertiges Image: `ghcr.io/uuuxy/schadenmeldung`.
- Caddy: Abschnitt aus `Caddyfile.beispiel`. Nur im Schulnetz erreichbar: dort `SCHULNETZ` einschalten.
- Zeitplan: `update.sh` plant beim Aufspielen die nächtliche Sicherung (2:30) und den Wächter (alle
  fünf Minuten) in der crontab von root ein, falls sie fehlen. Der Wächter startet Hängendes neu und mailt an `WAECHTER_AN`. Die Sicherung hält sieben Tage und
  kopiert mit `SICHERUNG_KOPIE` auf ein zweites Gerät. Probe: `./waechter.sh --probe` und
  `./sicherung-pruefen.sh`.
- Protokoll: App, Sicherung und Wächter schreiben ins Journal des Servers, `journalctl -t schadenmeldung`
  (braucht systemd, das prüft `update.sh`). Namen und Freitexte stehen nicht darin.
- Vor dem Start: Die Schul-IT nimmt den Server beim Mailserver von fail2ban aus (`ignoreip`), sonst
  sperren ein paar falsche Passwörter alle aus.

## Einstellungen (`.env`)

| Variable | Bedeutung | Vorgabe |
|---|---|---|
| `DATABASE_URL` | PostgreSQL; im Compose-Stack aus `POSTGRES_PASSWORD` gebaut | **Pflicht** |
| `IMAP_HOST`, `IMAP_PORT` | Mailserver der Schule für die Anmeldung, Name wie im Zertifikat; `mock` nur lokal | **Pflicht**, `993` |
| `ANMELDE_DOMAINS` | erlaubte Maildomains; bei genau einer genügt der Name vor dem @ | alle |
| `ALIAS_DOMAINS` | Kurzdomains derselben Postfächer, umgeschrieben auf `ANMELDE_DOMAINS` | – |
| `ADMINS` | feste Admins; weitere Rollen vergeben sie in der App | – |
| `SMTP_HOST`, `SMTP_PORT` | Mailserver für den Versand; ohne Host keine Mails | –, `587` |
| `SMTP_BENUTZER`, `SMTP_PASSWORT`, `MAIL_ABSENDER` | festes Postfach für den Versand; leer, wenn es unter Einstellungen → Mail eingetragen wird | – |
| `SCHLUESSEL` | verschlüsselt das in der App eingetragene Postfach; `update.sh` legt ihn an | – |
| `OEFFENTLICHE_ADRESSE` | Adresse der App, für die Knöpfe in der Mail | Pflicht mit SMTP |
| `WAECHTER_AN`, `SICHERUNG_KOPIE` | Empfänger der Warnungen; zweites Gerät für die Sicherung | erster Admin, – |
| `SCHULNAME` | Name auf Anmeldung und Datenschutzseite | – |
| `FOTO_TAGE`, `AUFBEWAHRUNG_TAGE` | Fotos bzw. Meldungen so viele Tage nach Erledigung löschen | `30`, `365` |
| `ZEITZONE` | für Uhrzeiten in Mail und Export | `Europe/Berlin` |
| `PROXY_VERTRAUEN` | `1` hinter Caddy: Absenderadresse aus `X-Real-IP` | aus |
| `PORT`, `APP_ENV` | Port; `local` nur auf dem Entwicklungsrechner | `8090`, `production` |

## Entwicklung

```bash
docker compose -f docker-compose.local.yml up -d          # PostgreSQL 5435, Mailpit 1025/8025
cd frontend && npm ci && npm run build && cd ..
APP_ENV=local IMAP_HOST=mock ANMELDE_DOMAINS=schule.de ADMINS=admin@schule.de \
  DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' \
  SMTP_HOST=localhost SMTP_PORT=1025 MAIL_ABSENDER=schaden@schule.de \
  OEFFENTLICHE_ADRESSE=http://localhost:8090 go run .
```

Dann `http://localhost:8090`; jedes Passwort gilt, Mails landen unter `http://localhost:8025`.

Tests: `TEST_DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' go test ./...`.
Die Browserprüfungen in `frontend/e2e/` (`npm run pruefen:…`) laufen gegen eine Wegwerf-Datenbank:
`docker exec schaden-db-local createdb -U schaden schaden_e2e`, den Server darauf starten, dann
`frontend/e2e/testdaten.sh`. Arbeitsregeln: [CLAUDE.md](CLAUDE.md).
