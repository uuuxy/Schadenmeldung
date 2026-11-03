# Übergabe: Schadenmeldung (Stand 2. November 2026, abends)

Für die nächste Claude-Code-Sitzung. Zuerst `CLAUDE.md` (Arbeitsregeln) und `README.md` lesen; diese
Datei sagt, wo wir stehen, was entschieden ist und was als Nächstes kommt.

## Worum es geht

Schadenmeldung für eine Gesamtschule in Hessen. Lehrkräfte melden Schäden
am Handy, der Hausmeister bearbeitet sie und bekommt Mails. Verlinkt wird die App aus LANiS
(Schulportal Hessen); dort gibt es keine Anmeldung, die man weiterreichen könnte, deshalb meldet die
App selbst über das Schul-Postfach an (IMAP), wie die Bibliothek und schul-orga des Nutzers.

**Der Nutzer:** baut die Schul-Apps selbst mit Claude Code (`~/Developer/Bibliothek`, `schul-orga`,
`inventur` …). Er diktiert, rechne also mit Hörfehlern („LANES“ = LANiS). Antworten auf Deutsch, per
du. Er will Nachweise statt Behauptungen: messen, testen, Screenshots ansehen.

**Seine Vorgaben, alle schon umgesetzt:**
- **Design:** Material 3 genau wie die Bausteine der Bibliothek
  (`~/Developer/Bibliothek/frontend/src/lib/components/ui/`), etwa der M3-Switch für „Dunkles Design“.
- **Handy zuerst:** Dark Mode, fast nichts tippen, das Datum setzt die App.
- **Kollegium:** sieht nur das Meldeformular, sonst nichts. Ohne QR-Code muss es genauso einfach gehen.
- **Hausmeister:** nicht stressen.
- **Datenbank:** PostgreSQL, ausdrücklich kein SQLite.
- **Tippflächen:** am Handy wirklich leicht zu treffen, mindestens 48 px, nachgemessen.

## Stand

Läuft unter **https://schaden.example.org** auf dem Hetzner (Testserver neben Bibliothek und
schul-orga), mit zwei festen Adressen der Verwaltung in `ADMINS` (nur in der Server-`.env`) und
ohne Mailversand. Die Anmeldung mit dem
Schul-Postfach klappt: Am 2. November kamen dort zwei Anmeldungen mit Adressen der Schule an.

### Server

Wo der Stack liegt, wie Caddy eingehängt ist und welche Fallen es gibt, steht nur lokal in
`SERVER.lokal.md` (von Git ausgenommen): Das Repo soll öffentlich werden können, ohne dass darin
Adressen, Zugänge oder Pfade des Servers stehen. Fehlt die Datei, den Nutzer fragen.

- **Git:** auf GitHub, `https://github.com/uuuxy/Schadenmeldung` (seit 3. November, vorher `uuuxy/schaden`),
  Zweig `main`, öffentlich, Lizenz EUPL-1.2. Die Geschichte ist bereinigt: keine Server-Angaben,
  keine Schuldaten, Beispielnamen wie `schule.example`. Image: `ghcr.io/uuuxy/schadenmeldung`. Commit-Nachrichten
  auf Deutsch, ohne Werkzeug-Hinweis, wie in der Bibliothek.
- **Technik:** Go (`net/http`, pgx) mit eingebetteter Oberfläche (Svelte 5 + Tailwind 4,
  `frontend/dist`) und eingebetteten Migrationen (`migrations/001_grundlage.sql`,
  `002_benachrichtigung.sql`). PostgreSQL 18, Docker Compose hinter dem vorhandenen Caddy (Netz
  `caddy_global_net`), distroless-Abbild mit 23 MB.
- **Rollen** (Tabelle `rollen`, gepflegt in der App unter „Einstellungen → Personen“):
  - `kollegium`: Vorgabe für jedes Postfach der Domain, sieht nur das Formular.
  - `hausmeister`, `leitung`, `admin`: dürfen alles, auch Räume, Arten und Personen; der Hausmeister
    bekommt die Mails.
  - `verwaltung` (Sekretariat): sieht die Übersicht, öffnet Meldungen, exportiert und druckt
    QR-Aufkleber, ändert nichts.
  - `ADMINS` (env) ist feste Verwaltung, damit sich niemand aussperrt. Der Nutzer rechnet mit bis zu 10 Personen im Backend.
- **Stati:** `offen`, `gesehen`, `bestellt`, `fachfirma`, `erledigt` (CHECK in Migration 002).
  „Gesehen“ ersetzt „In Arbeit“: Der Hausmeister zeigt damit, dass er Bescheid weiß, ohne eine Zusage, wann.
- **Mail an den Hausmeister** (`benachrichtigung.go`):
  - Jede neue Meldung geht sofort als eigene Mail hinaus; eine Sammelmail gibt es nicht (Entscheidung des Nutzers, 3. November). `mail_versand` ist der Postausgang (`postausgang.go`): Was scheitert, schickt der Server einen Tag lang alle 5 Minuten nach, und nichts kommt doppelt.
  - Die Knöpfe „Gesehen“, „Fachfirma“ und „Erledigt“ in der Mail öffnen `/aktion/<token>`. Das funktioniert ohne Anmeldung, 14 Tage lang, und erst ein Tipp auf der Seite löst etwas aus, damit Virenscanner beim Vorab-Öffnen nichts drücken.
  - Ein- und ausschalten lässt sich das je Person im Kontomenü.
- **Offline:** Service Worker (`frontend/public/sw.js`); der Ausgang in IndexedDB reicht Meldungen aus dem Funkloch nach.
- **Fotos:** liegen als `bytea` in Postgres und werden verkleinert, die Metadaten fallen weg.
- **Löschfristen:** in `aufraeumen.go`.
- **QR-Codes:** `/r/<code>`. Angemeldet wird nur einmal je Gerät (180 Tage, gleitend). Danach ist der Raum sofort gewählt.

### Geprüft

- 21 Go-Tests gegen Postgres, jeder Test mit eigenem Schema.
- Browserprüfungen im Handyformat (`frontend/e2e/`):
  - Der Ablauf QR → Anmeldung → Raum erkannt.
  - Das Kollegium sieht nur das Formular, die Leitung nur lesend.
  - Ein Mail-Knopf erledigt die Meldung ohne Anmeldung.
  - Alle Tippflächen messen mindestens 48 px. Tipps 3–6 px neben die Ränder treffen trotzdem.
  - Live: Eine neue Meldung erscheint beim Hausmeister nach rund 0,2–0,4 s, „Gesehen“ bei der
    Leitung genauso, ohne Neuladen; das Kollegium öffnet keinen Strom.
- Mailversand über Mailpit; Docker-Build und Compose-Probe.
- Handy-Screenshots für den Nutzer: https://claude.ai/artifact/MmpdCqftLEHp4qr2hc7t75 (privat), vorher
  und nachher der M3-Durchsicht: https://claude.ai/artifact/UMRmh6j4WRnX1F9QyNm73R (privat).

## Mailserver der Schule

Einstellungen und Prüfungen zum Mailserver der Schule (IMAP und SMTP) stehen ebenfalls
nur in `SERVER.lokal.md`.

## Als Nächstes (Reihenfolge)

1. **Rollen vergeben:** Die Verwaltung trägt Hausmeister und Leitung in der App ein („Räume &
   Personen“). Weitere feste Verwaltung: `ADMINS` in der Server-`.env`, dann `docker compose up -d`
   (Sitzungen bleiben). **Nie selbst mit erfundenen Zugangsdaten gegen den Schulserver anmelden:**
   Fehlversuche können Konten sperren.
2. **Erledigt am 2. November (abends):** Export als PDF (Liste wie angezeigt, nach Bereichen, mit
   Fotos) und Excel (alle gespeicherten Meldungen, ohne Namen); Arten unter Einstellungen an- und
   ausschaltbar, Medientechnik zu Beginn aus (Rückmeldung: die IT hat ein eigenes Ticketsystem);
   Räume umbenennen ohne neuen QR-Code; M3-Durchsicht aller Seiten nach m3.material.io (Übersicht
   als schlanke Liste mit Detailseite, Formular schrumpft Frage für Frage, Rückfrage-Dialog statt
   `confirm()`, Snackbar mit Rückgängig bleibt stehen). Ein Git-Hook vor jedem Commit
   (gofmt, vet, Größe) ist entworfen, aber nicht eingerichtet: Er braucht die Erlaubnis des Nutzers.
3. **Absender-Postfach** für den Versand bei der Schul-IT anfragen, etwa `schadenmeldung@schule.example`.
   `MAIL_ABSENDER` muss dessen eigene Adresse sein; Server lassen meist nur die angemeldete Adresse als Absender zu.
4. Optional: IMAP mit STARTTLS (Port ≠ 993) unterstützen, falls 993 einmal wegfällt.
   go-imap v1 kann `c.StartTLS`; LOGIN nie ohne TLS schicken. Test mit dem Server aus go-imap v1
   (`server` + `backend/memory`, selbstsigniertes Zertifikat) und einem Haken für `RootCAs`.
5. Auf einem echten Handy testen, jetzt über https://schaden.example.org (auch Offline und Homescreen).

## Offene Entscheidungen des Nutzers (nicht vorwegnehmen)

- „Gesehen“ statt „In Arbeit“ bestätigen.
- Hat der Hausmeister ein Schulpostfach? Vorschlag: Funktionspostfach `hausmeister@…`.
- Haben Schüler Postfächer unter derselben Domain? Dann könnten sie melden. Gewollt?
- Offen im Internet mit Anmeldung (empfohlen, Vorgabe) oder nur im Schulnetz (Caddy `SCHULNETZ`)?
- Push-Benachrichtigung statt oder zusätzlich zur Mail (nicht gebaut).
- Organisatorisch: Verzeichnis der Verarbeitungstätigkeiten, Datenschutzbeauftragte(r), Personalrat.
  Der Hausmeister ist meist beim Schulträger angestellt, dann ist dessen Personalrat zuständig.

## Lokal arbeiten

```bash
docker compose -f docker-compose.local.yml up -d     # Postgres :5435, Mailpit :1025/:8025 (Bibliothek nutzt :5434 – nicht anfassen)
TEST_DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' go test ./...
cd frontend && npm run build                          # 0 Warnungen erwartet
```

Testserver mit Mock-Anmeldung, jedes Passwort gilt:

```bash
APP_ENV=local IMAP_HOST=mock DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' \
  ANMELDE_DOMAINS=schule.de SCHULNAME=Musterschule ADMINS=admin@schule.de \
  SMTP_HOST=localhost SMTP_PORT=1025 MAIL_ABSENDER=schaden@schule.de OEFFENTLICHE_ADRESSE=http://127.0.0.1:8090 go run .
```

Browserprüfungen (Chromium von Playwright ist auf dem Mac installiert):
- `frontend/e2e/testdaten.sh` füllt eine **frische** Datenbank. Davor die DB leeren:
  `docker exec schaden-db-local psql -U schaden -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'`,
  dann den Server neu starten. Mailpit leeren: `curl -X DELETE http://127.0.0.1:8025/api/v1/messages`.
- `npm run pruefen:ablauf` prüft QR, Rollen und Mail-Knopf und speichert Screenshots. Es braucht die
  frischen Testdaten, sonst schlagen die Mail-Prüfungen fehl.
- `npm run pruefen:tippflaechen` misst alle Tippflächen; `npm run pruefen:tippen` tippt neben die Ränder.
- `npm run pruefen:live` misst, wie schnell die Übersicht eine neue Meldung zeigt (Testdaten nötig).

## Fallstricke

- Radien und Schrift folgen der M3-Skala in `frontend/src/styles/mass.css`, nicht dem Tailwind-Standard:
  `rounded-2xl` = 28 px, `text-title-medium` gibt es.
- Kleine Bauteile tragen `.treffer`. Kein `overflow-hidden` darum, sonst greift die 48-px-Fläche nicht.
  Chip-Gruppen haben `gap-y-4`.
- Arten und Stati stehen an mehreren Stellen: `server.go`, `frontend/src/lib/kategorien.js`,
  `status.js` und der CHECK der Migration.
- Texte über den Hausmeister neutral halten, ohne „er“ oder „sein“.
- macOS kennt kein `timeout`; dafür `perl -e 'alarm 10; exec @ARGV' …`. LibreSSL-`openssl` kann kein `-brief`.
- zsh: `--include=*.svelte` in Anführungszeichen setzen. Nach Python-Änderungen an Go-Dateien `gofmt -w .` laufen lassen.
