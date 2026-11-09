# Schadenmeldung

Das Kollegium meldet Schäden im Schulgebäude in wenigen Sekunden vom Handy, der Hausmeister hakt sie
ab. Statt einer Tabelle gibt es große Kacheln und Knöpfe zum Antippen; getippt wird fast nichts.

## Was es kann

- **Wo? – mit oder ohne QR-Code:** Der QR-Aufkleber an der Tür füllt den Raum aus, gescannt mit der
  Kamera-App oder mit dem Symbol rechts in „Raum suchen“, dann bleibt man in der App. Ohne Code genügt
  ein Tipp auf „Zuletzt“, zwei Tipps über Bereich → Raum, oder kurz „204“ ins Suchfeld. Für alles
  ohne Raum gibt es „Anderer Ort“.
- **Was? – antippen statt tippen:** Art (Licht, Heizung, Wasser …) → typischer Fall („Lampe defekt“)
  → Melden. Jede beantwortete Frage schrumpft auf eine Zeile mit „Ändern“, so steht die offene Frage
  immer oben im Bild. Datum, Uhrzeit und wer gemeldet hat, setzt der Server. Foto aus der Kamera,
  Notiz und „Dringend / Gefahr“ sind freiwillig. Tippt jemand zu früh auf „Melden“, wird die offene
  Frage rot und sagt, was fehlt.
- **Keine Doppelmeldungen:** Im gewählten Raum steht sofort, was dort schon gemeldet ist und wie
  weit es ist.
- **Rollen:** Das Kollegium sieht nur „Schaden melden“, sonst nichts. **Hausmeister**, **Leitung**
  und **Admin** dürfen alles: Meldungen bearbeiten und löschen, Räume, Arten, Personen und Rollen. Die
  **Verwaltung** (Sekretariat) sieht die Übersicht und druckt QR-Aufkleber, ändert aber nichts. Die
  Adressen aus `ADMINS` sind immer Admin, damit sich niemand aussperren kann.
- **Bearbeiten:** eine schlanke Liste der offenen Meldungen, Dringendes zuerst, Filter nach Bereich
  (ein Chip, die Bereiche im Blatt). Hinten an jeder Zeile das Häkchen „Erledigt“ mit Rückgängig.
  Ein Tipp auf die Zeile öffnet die Meldung: Foto groß, Zwischenstand mit einem Tipp („In Arbeit“,
  „Bestellt“, „Fachfirma“), ein interner Vermerk (sehen alle mit Übersicht, nicht das Kollegium),
  Löschen mit Rückfrage. Ziehen zum Aktualisieren.
- **Drucken:** Auf der Seite einer Meldung druckt das Drucker-Symbol sie als PDF, eine A4-Seite mit großem
  Foto. Mehrere wählt man in der Liste wie in Gmail (lange drücken, dann antippen, oder „Drucken und
  exportieren“ → „Meldungen auswählen“); das PDF hat dann eine Seite je Meldung. Ohne Namen der Meldenden.
- **Export:** „Drucken und exportieren“ in der Übersicht lädt ein PDF der Liste, wie sie gerade steht (offen oder
  erledigt, auf Wunsch nur ein Bereich; nach Bereichen gegliedert, mit kleinen Fotos, zum Drucken oder
  für den Schulträger), oder eine Excel-Tabelle mit allen gespeicherten Meldungen (echte Datumszellen,
  feste Kopfzeile, Filter). Beide ohne Namen der Meldenden.
- **Zwei Ebenen, live:** Alle mit Übersicht wechseln oben in der Kopfleiste
  zwischen „Übersicht“ und „Melden“, denn auch sie melden Schäden. Die Übersicht aktualisiert sich
  sofort, wenn jemand meldet oder einen Stand setzt (Server-Sent Events), ohne Neuladen.
- **Farbe je Art:** Licht, Heizung, Wasser … leuchten beim Antippen in ihrer eigenen, gedämpften
  Farbe (M3 „custom colors“, zur Grundfarbe hin harmonisiert); dieselbe Farbe trägt die Zeile in der
  Übersicht.
- **Mail an den Hausmeister:** Jede neue Meldung kommt sofort als eigene Mail, Dringendes mit
  „Dringend“ im Betreff, jede nur einmal. Antwortet der Mailserver nicht, schickt die App einen Tag
  lang alle fünf Minuten nach. In der Mail stehen die Knöpfe „In Arbeit“, „Fachfirma“ und „Erledigt“; sie wirken ohne Anmeldung, 14 Tage lang, und erst beim Antippen auf der Seite dahinter
  (Virenscanner, die Links vorab öffnen, lösen nichts aus). Jeder schaltet seine Mails im Kontomenü
  selbst an oder ab; außer dem Hausmeister bekommt niemand ohne eigene Wahl welche.
- **„In Arbeit“:** Ein Tipp zeigt dem Kollegium, dass sich jemand kümmert – ohne Zusage,
  wann. Angezeigt wird der Stand, nicht der Zeitpunkt.
- **Einstellungen** (Kontomenü): Räume per Liste einfügen (auch aus Excel oder CSV), umbenennen,
  ohne dass der QR-Aufkleber ungültig wird, QR-Aufkleber drucken (12 je A4-Blatt); Arten ein- und
  ausschalten (Medientechnik ist zu Beginn aus, weil die IT dafür ein eigenes Ticketsystem hat; die
  übrigen Kacheln rücken zusammen); Personen und Rollen.
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
| Lizenz | [EUPL-1.2](LICENSE) |

## Ausprobieren auf dem eigenen Rechner

```bash
docker compose -f docker-compose.local.yml up -d     # PostgreSQL (5435) und Mailpit (1025/8025)
cd frontend && npm ci && npm run build && cd ..
APP_ENV=local IMAP_HOST=mock \
  DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' \
  ANMELDE_DOMAINS=schule.de ADMINS=admin@schule.de \
  SMTP_HOST=localhost SMTP_PORT=1025 MAIL_ABSENDER=schaden@schule.de \
  OEFFENTLICHE_ADRESSE=http://localhost:8090 \
  go run .
```

Dann `http://localhost:8090` öffnen. Mit `IMAP_HOST=mock` gilt jedes Passwort (nur zusammen mit
`APP_ENV=local`, sonst startet der Server nicht). Als `admin` anmelden, unter „Einstellungen“
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
2. `./update.sh` im Projektordner, wie in der Bibliothek: Es sichert die Datenbank nach `backups/`
   (7 Tage), holt den Stand von `origin/main`, baut, startet und wartet, bis die App antwortet; geht
   etwas schief, nennt es den Weg zurück. Dasselbe für jedes spätere Update. Das fertige Image von
   GitHub (`.github/workflows/docker.yml`, erst nach grünen Prüfungen) ginge auch:
   `docker login ghcr.io -u uuuxy`, dann `docker compose pull && docker compose up -d`.
3. Den Abschnitt aus `Caddyfile.beispiel` in den Caddy übernehmen, Namen eintragen, `caddy reload`.
   Für den Namen braucht es einen DNS-Eintrag auf den Server; das Zertifikat holt Caddy selbst.
4. Als Admin anmelden, unter „Einstellungen“ Räume einfügen und den Hausmeister eintragen,
   dann die QR-Aufkleber drucken. Dafür die Seite über die echte Adresse öffnen, denn die Codes
   führen auf die Adresse, unter der sie erzeugt werden.
5. Wächter und nächtliche Sicherung einrichten (`crontab -e` als root, Pfad anpassen):

   ```
   */5 * * * * /pfad/zur/schadenmeldung/waechter.sh 2>&1 | logger -t schadenmeldung
   30 2 * * * /pfad/zur/schadenmeldung/sicherung.sh 2>&1 | logger -t schadenmeldung
   ```

   Der Wächter startet App oder Datenbank neu, wenn sie hängen, und mailt an `WAECHTER_AN` (sonst
   an den ersten aus `ADMINS`), wenn das nicht hilft, die Platte fast voll ist oder die Sicherung
   ausbleibt. `./waechter.sh --probe` schickt eine Probemail. Vor dem Start an der Schule:
   [GENERALPROBE.md](GENERALPROBE.md).

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
| `MAIL_ABSENDER` | Absender der Mails; leer, wenn das Postfach unter Einstellungen → Mail eingetragen wird | – |
| `SCHLUESSEL` | 64 Hex-Zeichen; verschlüsselt das Passwort eines in der App eingetragenen Postfachs, `update.sh` legt ihn an | – |
| `OEFFENTLICHE_ADRESSE` | Adresse der App für die Knöpfe in der Mail | Pflicht mit SMTP |
| `ZEITZONE` | für die Uhrzeiten in Mail und Export | `Europe/Berlin` |
| `SCHULNAME` | Name auf Anmeldung und Datenschutzseite | – |
| `FOTO_TAGE` | Fotos so viele Tage nach Erledigung löschen | `30` |
| `AUFBEWAHRUNG_TAGE` | Meldungen so viele Tage nach Erledigung löschen | `365` |
| `PROXY_VERTRAUEN` | `1` hinter Caddy: Absenderadresse aus `X-Real-IP` | aus |
| `PORT` | Port des Servers | `8090` |
| `APP_ENV` | `local` nur auf dem Entwicklungsrechner | `production` |

## Sicherung

Alles liegt in PostgreSQL, auch die Fotos. Jede Nacht sichert `sicherung.sh` die Datenbank nach `backups/` und
löscht Sicherungen nach sieben Tagen. Einmal auf dem Server einrichten (`crontab -e` als root):

```
30 2 * * * /pfad/zur/schadenmeldung/sicherung.sh 2>&1 | logger -t schadenmeldung
```

Ob es geklappt hat, zeigt `journalctl -t schadenmeldung`. Eine Sicherung auf demselben Server hilft nicht, wenn
dessen Platte ausfällt: Mit `SICHERUNG_KOPIE` in `.env` (ein eingehängtes NAS oder `benutzer@nas:/pfad` mit
SSH-Schlüssel) spiegelt `sicherung.sh` den Ordner jede Nacht auf ein zweites Gerät, samt Löschen nach sieben
Tagen. Ob sich eine Sicherung zurückspielen lässt, zeigt `./sicherung-pruefen.sh`: Es spielt die neueste
probeweise in eine eigene Datenbank ein und zählt Meldungen und Räume, ohne die echte anzufassen.

Von Hand ist eine Sicherung ein einziger Befehl:

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
- Wer gemeldet hat, sehen nur Hausmeister, Leitung, Verwaltung und Admins (für Rückfragen); in den
  Export (PDF, Excel) kommt der Name nicht, und auch nicht, wann „In Arbeit“ gesetzt wurde.
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
`npm run pruefen:tippflaechen` (jede Tippfläche ≥ 48 px, auch in Blättern und Menüs, ohne
Überlappung), `npm run pruefen:tippen` (Tipps neben die Ränder) und `npm run pruefen:live` (neue
Meldung und neuer Stand erscheinen ohne Neuladen, unter 2 Sekunden). `ablauf` prüft außerdem Liste,
Detailseite, Häkchen mit Rückgängig, Räume umbenennen und Arten an- und ausschalten.

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
| `benachrichtigung.go`, `postausgang.go` | Mail an den Hausmeister: jede Meldung sofort, Nachholen bei Ausfall, Knöpfe in der Mail |
| `aufraeumen.go` | Löschfristen |
| `einstellungen.go` | Einstellungen für die Oberfläche, Namen der Arten und Stati, Arten an und aus |
| `export.go`, `export_pdf*.go` | Export als Excel-Tabelle und PDF (gofpdf wie in der Bibliothek) |
| `druck.go` | Einzelne Meldungen als PDF zum Drucken, je eine Seite mit Foto |
| `xlsx.go`, `xlsx_schreiben.go` | Excel lesen (Raumimport) und schreiben, ohne fremde Bibliothek |
| `frontend/` | Svelte 5, Tailwind 4, Material 3 (Rollen in `src/styles/rollen.css`, Bausteine in `src/components/ui/`) |
| `frontend/public/sw.js` | Service Worker (Start ohne Netz) |
| `frontend/src/lib/ausgang.js` | Ausgang für Meldungen aus dem Funkloch (IndexedDB) |
