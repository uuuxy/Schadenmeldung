# Schadenmeldung

Das Kollegium meldet Schäden im Schulgebäude in wenigen Sekunden vom Handy, der Hausmeister hakt sie
ab. Statt einer Tabelle gibt es große Kacheln und Knöpfe zum Antippen; getippt wird fast nichts.

## Was es kann

- **Wo? – mit oder ohne QR-Code:** Der QR-Aufkleber an der Tür füllt den Raum aus. Ohne Code genügt
  ein Tipp auf „Zuletzt“, zwei Tipps über Bereich → Raum, oder kurz „204“ ins Suchfeld. Für alles
  ohne Raum gibt es „Anderer Ort“.
- **Was? – antippen statt tippen:** Art (Licht, Heizung, Wasser …) → typischer Fall („Lampe defekt“)
  → Melden. Datum, Uhrzeit und wer gemeldet hat, setzt der Server. Foto aus der Kamera, Notiz und
  „Dringend / Gefahr“ sind freiwillig.
- **Keine Doppelmeldungen:** Im gewählten Raum steht sofort, was dort schon gemeldet ist und wie
  weit es ist.
- **Rollen:** Das Kollegium sieht nur „Schaden melden“, sonst nichts. **Hausmeister** bearbeiten die
  Aufgaben, die **Leitung** (z. B. Schulleitung) sieht alles und ändert nichts, die **Verwaltung**
  darf alles, auch Räume und Personen. Rollen vergibt die Verwaltung in der App; die Adressen aus
  `ADMINS` sind immer Verwaltung, damit sich niemand aussperren kann.
- **Hausmeister:** offene Aufgaben, Dringendes zuerst, Filter nach Bereich; Zwischenstand mit einem
  Tipp („Gesehen“, „Bestellt“, „Fachfirma“), Antwort mit Schnelltexten, „Erledigt“ mit Rückgängig,
  Ziehen zum Aktualisieren.
- **Zwei Ebenen, live:** Hausmeister, Leitung und Verwaltung wechseln oben in der Kopfleiste
  zwischen „Übersicht“ und „Melden“, denn auch sie melden Schäden. Die Übersicht aktualisiert sich
  sofort, wenn jemand meldet oder einen Stand setzt (Server-Sent Events), ohne Neuladen.
- **Farbe je Art:** Licht, Heizung, Wasser … leuchten beim Antippen in ihrer eigenen, gedämpften
  Farbe (M3 „custom colors“, zur Grundfarbe hin harmonisiert); dieselbe Farbe trägt die Karte in der
  Übersicht.
- **Mail an den Hausmeister, ohne Stress:** Dringendes sofort, alles andere als Sammelmail Mo–Fr um
  7 und 12 Uhr (einstellbar), jede Meldung nur einmal. In der Mail stehen die Knöpfe „Gesehen“ und
  „Erledigt“; sie wirken ohne Anmeldung, 14 Tage lang, und erst beim Antippen auf der Seite dahinter
  (Virenscanner, die Links vorab öffnen, lösen nichts aus). Jeder schaltet seine Mails im Kontomenü
  selbst an oder ab; Leitung und Verwaltung bekommen ohne eigene Wahl keine.
- **„Gesehen“ statt „In Arbeit“:** Ein Tipp zeigt dem Kollegium, dass Bescheid gesagt ist – ohne
  Zusage, wann. Angezeigt wird der Stand, nicht der Zeitpunkt.
- **Verwaltung:** Räume per Liste einfügen (auch aus einer Tabelle), QR-Aufkleber drucken
  (12 je A4-Blatt), Personen und Rollen.
- **Funkloch-fest:** Die App startet sofort, auch ohne Netz (Service Worker), zeigt den letzten
  Stand, und Meldungen aus dem Keller gehen samt Foto automatisch raus, sobald wieder Netz da ist.
- **Anmeldung mit dem Schul-Postfach** (IMAP, wie Bibliothek und schul-orga). Das Handy bleibt
  angemeldet; das Passwort wird nirgends gespeichert.
- **Material 3** mit denselben Bausteinen wie die Bibliothek (Button, Switch, Filter-Chips, Segmente,
  Suchpille, Feld, Reiter, Status-Chip in `frontend/src/components/ui/`), für das Handy gebaut.
  „Dunkles Design“ ist ein Schalter im Kontomenü: ohne Wahl wie das System, sonst je Gerät. Als App
  auf den Homescreen legbar, mit Kurzbefehlen am App-Symbol.

## Technik

| | |
|---|---|
| Server | Go, `net/http`, pgx; Oberfläche und Migrationen sind eingebettet (ein Programm) |
| Datenbank | PostgreSQL 18, nummerierte Migrationen in `migrations/`, laufen beim Start; Fotos als `bytea` |
| Oberfläche | Svelte 5 (Runes), Tailwind 4, Vite; Material 3 mit den Farbrollen der Bibliothek |
| Offline | eigener Service Worker (`frontend/public/sw.js`), Ausgang in IndexedDB |
| Anmeldung | IMAP gegen den Schul-Mailserver |
| Betrieb | Docker Compose hinter Caddy |

## Ausprobieren auf dem eigenen Rechner

```bash
docker compose -f docker-compose.local.yml up -d     # PostgreSQL (5435) und Mailpit (1025/8025)
cd frontend && npm ci && npm run build && cd ..
APP_ENV=local IMAP_HOST=mock \
  DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' \
  ANMELDE_DOMAINS=schule.de ADMINS=verwaltung@schule.de \
  SMTP_HOST=localhost SMTP_PORT=1025 MAIL_ABSENDER=schaden@schule.de \
  OEFFENTLICHE_ADRESSE=http://localhost:8090 \
  go run .
```

Dann `http://localhost:8090` öffnen. Mit `IMAP_HOST=mock` gilt jedes Passwort (nur zusammen mit
`APP_ENV=local`, sonst startet der Server nicht). Als `verwaltung` anmelden, unter „Räume & Personen“
Räume einfügen und z. B. `hausmeister` die Rolle Hausmeister geben; jeder andere Name meldet. Die
Mails an den Hausmeister landen in Mailpit unter `http://localhost:8025`. Ein Handy im selben WLAN erreicht die App unter
`http://<IP des Rechners>:8090`; der Offline-Betrieb braucht https und läuft deshalb erst auf dem
Server (oder unter `localhost`).

Oberfläche mit sofortigem Neuladen: `cd frontend && npm run dev` (leitet `/api` an Port 8090 weiter).

## Betrieb auf dem Server

Wie die Bibliothek: Docker Compose hinter dem vorhandenen Caddy im Netz `caddy_global_net`; die
Datenbank hängt nur am internen Netz des Stacks.

1. `.env` aus `.env.example` anlegen: Datenbankpasswort (`openssl rand -hex 24`), Mailserver für
   die Anmeldung (IMAP) und für den Versand (SMTP), Domain, feste Verwaltung (`ADMINS`).
2. `docker compose up -d --build`
3. Den Abschnitt aus `Caddyfile.beispiel` in den Caddy übernehmen, Namen eintragen, `caddy reload`.
   Für den Namen braucht es einen DNS-Eintrag auf den Server; das Zertifikat holt Caddy selbst.
4. Als Verwaltung anmelden, unter „Räume & Personen“ Räume einfügen und den Hausmeister eintragen,
   dann die QR-Aufkleber drucken. Dafür die Seite über die echte Adresse öffnen, denn die Codes
   führen auf die Adresse, unter der sie erzeugt werden.

**Erreichbarkeit:** Ohne weitere Einstellung ist die App überall erreichbar, geschützt durch die
Anmeldung, also auch über mobile Daten im Keller oder von zu Hause. Soll sie nur im Schulnetz
laufen, schaltet der Caddy-Abschnitt das mit `SCHULNETZ` ein (zwei Zeilen, siehe Beispiel).

## Einstellungen

| Variable | Bedeutung | Vorgabe |
|---|---|---|
| `DATABASE_URL` | PostgreSQL; im Compose-Stack aus `POSTGRES_PASSWORD` gebaut | **Pflicht** |
| `IMAP_HOST` | Mailserver der Schule für die Anmeldung; der Name muss im Zertifikat des Servers stehen; `mock` nur lokal | **Pflicht** |
| `IMAP_PORT` | IMAP über TLS | `993` |
| `ANMELDE_DOMAINS` | erlaubte Maildomains, durch Komma getrennt; bei genau einer genügt der Name vor dem @ | alle, die der Mailserver kennt |
| `ALIAS_DOMAINS` | Kurzdomains derselben Postfächer, etwa `kurz.example`; die App schreibt sie beim Anmelden auf die eine Domain aus `ANMELDE_DOMAINS` um | – |
| `ADMINS` | feste Verwaltung, auch als Name oder in Kurzform; weitere Rollen vergibt sie in der App | – |
| `SMTP_HOST`, `SMTP_PORT` | Mailserver für den Versand an den Hausmeister; Name wie im Zertifikat; ohne Host keine Mails | –, `587` |
| `SMTP_BENUTZER`, `SMTP_PASSWORT` | Anmeldung am Mailserver (nur über TLS) | – |
| `MAIL_ABSENDER` | Absender der Mails | Pflicht mit SMTP |
| `OEFFENTLICHE_ADRESSE` | Adresse der App für die Knöpfe in der Mail | Pflicht mit SMTP |
| `SAMMELMAIL_ZEITEN` | Uhrzeiten der Sammelmail (Mo–Fr), `aus` schaltet sie ab | `07:00,12:00` |
| `ZEITZONE` | für Sammelmail und Uhrzeiten in der Mail | `Europe/Berlin` |
| `SCHULNAME` | Name auf Anmeldung und Datenschutzseite | – |
| `FOTO_TAGE` | Fotos so viele Tage nach Erledigung löschen | `30` |
| `AUFBEWAHRUNG_TAGE` | Meldungen so viele Tage nach Erledigung löschen | `365` |
| `PROXY_VERTRAUEN` | `1` hinter Caddy: Absenderadresse aus `X-Real-IP` | aus |
| `PORT` | Port des Servers | `8090` |
| `APP_ENV` | `local` nur auf dem Entwicklungsrechner | `production` |

## Sicherung

Alles liegt in PostgreSQL, auch die Fotos. Eine Sicherung ist deshalb ein einziger Befehl:

```bash
docker compose exec -T db pg_dump -U schaden -Fc schaden > sicherung-$(date +%F).dump
# zurückspielen: docker compose exec -T db pg_restore -U schaden -d schaden --clean < sicherung-….dump
```

## Datenschutz

Eingebaut:

- Keine fremden Dienste, keine CDNs, kein Tracking, kein Cookie-Banner nötig (nur das
  Anmelde-Cookie, technisch notwendig).
- Das Passwort geht nur an den Mailserver; in der Datenbank steht nur der Hash des Sitzungsschlüssels.
- Fotos werden verkleinert und neu geschrieben, dabei fallen GPS-Ort und Gerätedaten weg.
- Wer gemeldet hat, sehen nur Hausmeister und Verwaltung (für Rückfragen).
- Fehlversuchsbremse bei der Anmeldung: 5 Fehlversuche je Konto bzw. 20 je Adresse in 15 Minuten.
  Ein Ausfall des Mailservers zählt nicht als Fehlversuch.
- Löschfristen laufen stündlich von selbst. Die Datenschutzseite in der App nennt sie.
- Auf dem Handy liegt für den Offline-Betrieb nur der letzte Stand der Liste und der Ausgang; beides
  löscht das Abmelden.

Organisatorisch offen: Eintrag ins Verzeichnis der Verarbeitungstätigkeiten, die oder den
Datenschutzbeauftragten informieren, den Personalrat einbinden (die Bearbeitungszeiten des
Hausmeisters sind Arbeitsdaten einer Person; Hausmeister sind meist beim Schulträger angestellt).
Die Entwürfe der Bibliothek (`docs/datenschutz/`) lassen sich dafür kürzen.

## Tests

```bash
docker compose -f docker-compose.local.yml up -d
TEST_DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' go test ./...
```

Jeder Test bekommt ein eigenes Schema und räumt es wieder ab. Ohne `TEST_DATABASE_URL` schlagen die
Datenbanktests fehl, statt still grün zu überspringen; wer bewusst ohne testet, setzt `OHNE_DB=1`.

Browserprüfungen im Handyformat liegen in `frontend/e2e/` (Playwright): `npm run pruefen:ablauf`
(QR, Rollen, Mail-Knopf; braucht frische Testdaten aus `e2e/testdaten.sh` und Mailpit),
`npm run pruefen:tippflaechen` (jede Tippfläche ≥ 48 px), `npm run pruefen:tippen` (Tipps neben die Ränder)
und `npm run pruefen:live` (neue Meldung und neuer Stand erscheinen ohne Neuladen, unter 2 Sekunden).

## Aufbau

| Datei | Inhalt |
|---|---|
| `main.go` | Start, Unterbefehl `gesund` für den Docker-Gesundheitscheck |
| `konfiguration.go` | Umgebungsvariablen, Rollen, Domainprüfung |
| `anmeldung.go` | IMAP-Anmeldung, Fehlversuchsbremse, Sitzungen |
| `server.go` | Routen, Schutzköpfe, Meldungen, Räume, Fotos, Oberfläche |
| `speicher.go` | PostgreSQL (pgx) und Migrationen |
| `migrations/` | nummerierte SQL-Dateien, laufen beim Start in eigener Transaktion |
| `fotos.go` | Fotos prüfen, verkleinern, Metadaten entfernen |
| `rollen.go` | Rollen: wer mehr sieht als das Meldeformular |
| `benachrichtigung.go` | Mail an den Hausmeister: sofort, Sammelmail, Knöpfe in der Mail |
| `aufraeumen.go` | Löschfristen |
| `frontend/` | Svelte 5, Tailwind 4, Material 3 (Rollen in `src/styles/rollen.css`, Bausteine in `src/components/ui/`) |
| `frontend/public/sw.js` | Service Worker (Start ohne Netz) |
| `frontend/src/lib/ausgang.js` | Ausgang für Meldungen aus dem Funkloch (IndexedDB) |
