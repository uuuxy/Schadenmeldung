# Arbeitsregeln für dieses Projekt

Die Schadenmeldung soll schlank bleiben: ein Go-Programm mit eingebetteter Svelte-Oberfläche und
PostgreSQL, keine weiteren Dienste. Bevor etwas Neues dazukommt: Braucht es das wirklich?

## Oberfläche

- **Material 3, Handy zuerst.** Ohne Präfix gestaltet man für das Handy, `sm:`/`md:` für größere
  Bildschirme. Eingabefelder mindestens 16 px Schrift (sonst zoomt iOS).
- **Jede Tippfläche trifft auf 48 px** (M3 „touch target“). Was sichtbar kleiner ist (Chip 32 px,
  Knopf 40 px), trägt `.treffer` (komponenten.css), Symbolknöpfe `.icon-btn`. Chip-Gruppen haben
  `gap-y-4`, damit sich die Flächen nicht überlappen, und kein `overflow-hidden` um kleine Bauteile,
  sonst wird die Fläche abgeschnitten. Nachmessen: Abstände und Größen im Handyformat (390 px).
- **Bausteine statt Handarbeit:** Knöpfe, Schalter, Chips, Segmente, Felder, Suche, Reiter und
  Status-Chips kommen aus `frontend/src/components/ui/`, gebaut wie die gleichnamigen Bausteine der
  Bibliothek (dort steht jeweils die M3-Begründung). Auswahl liegt auf secondary-container; nur die
  Kacheln der Arten leuchten in ihrer eigenen Farbe (`kat-*`, M3 „custom colors“, Fläche in
  Buntheit 24, damit es ruhig bleibt). Knöpfe tragen keinen Schatten, Chips brechen um statt
  seitlich zu scrollen.
- **Zwei Ebenen:** Wer die Übersicht sieht (Hausmeister, Leitung, Verwaltung), wechselt in der
  Kopfleiste per Segmented Button zwischen „Übersicht“ und „Melden“ (`KopfLeiste ebene=…`), nicht
  per Switch: M3 nimmt Switches nur für Ein/Aus. Die Übersicht aktualisiert sich live über
  `/api/ereignisse`; jeder neue Schreibweg für Meldungen ruft `a.verteiler.melde()`.
- **Liste und Detail (M3 list-detail):** Eine Zeile der Übersicht trägt bewusst nur eine Zusatzaktion
  (das Häkchen „Erledigt“), damit die Liste ruhig bleibt; alles Weitere steht auf der Seite der Meldung
  (`/meldung/<id>`). M3 erlaubt mehrere Aktionen je Zeile (Stand November 2026), mehr als drei Knöpfe
  nebeneinander sind aber zu viel. Endgültiges fragt über `frage()` nach (`lib/rueckfrage.svelte.js`),
  nie über `confirm()`.
- **Farben nur über Rollen** aus `frontend/src/styles/rollen.css` (`bg-surface`,
  `text-on-surface-variant`, `bg-primary-container` …), nie Palettenfarben. Jede Rolle hat einen
  Wert im hellen und im dunklen Schema. Einzige Ausnahme: die Papier-Rollen (`papier`, `auf-papier`,
  `papier-akzent` …) für den QR-Bogen bleiben fest, denn Papier ist weiß.
- **Schrift über die M3-Stufen** aus `mass.css` (`text-title-medium`, `text-body-large`,
  `text-label-large` …), Formen über die M3-Radien (`rounded-sm` = Chip, `rounded-xl` = Karte,
  `rounded-2xl` = Kachel/Blatt).
- **Symbole:** Material Symbols Rounded. Neues Symbol in `frontend/scripts/symbole.mjs` eintragen,
  dann `npm run symbole`.
- **Wenig tippen:** Neue Eingaben zuerst als Kacheln oder Chips denken, Freitext nur als Ausweg.
  Alles muss auch ohne QR-Code in wenigen Tipps gehen.
- **Funkloch:** Schreibende Wege dürfen im Funkloch nichts verlieren: Meldungen und Stand-Änderungen
  warten im Ausgang (`lib/ausgang.js`, `lib/nachreichen.js`), bis wieder Netz da ist, und
  `ladeMeldungen` legt wartende Änderungen über den Stand vom Server. Der Service Worker cacht nie
  `/api/` oder `/fotos/`.

## Rollen und Mail

- Kollegium sieht nur das Meldeformular; Hausmeister, Leitung und Admin dürfen alles, die
  Verwaltung (Sekretariat) sieht und druckt QR-Aufkleber, ändert aber nichts (`rollen.go`,
  `zustand.svelte.js`). Jede neue Tür prüft die Rolle im Server, nicht nur die Oberfläche.
- Mails und Mitteilungen aufs Handy sollen den Hausmeister nicht hetzen: eine je Meldung, keine
  Erinnerungen, keine Eskalation, keine Zeitstempel von „In Arbeit“ in der Oberfläche. Mail-Links lösen nie beim Öffnen etwas aus, nur beim Antippen.
- Wer meldet, bekommt eine Mail, sobald die Meldung erledigt ist: eine je Meldung, erst nach zehn
  Minuten (so lange greift „Rückgängig“), ohne den internen Vermerk; wer selbst erledigt, bekommt keine
  (`rueckmeldung.go`). Die Karte „Erledigt“ im Formular gibt es nur, wenn der Server keine Mail schickt.
- Lokal landen Mails in Mailpit (`docker-compose.local.yml`, http://localhost:8025).

## Datenbank

- PostgreSQL mit pgx. Schemaänderungen nur als neue, nummerierte Datei in `migrations/`; eine
  eingespielte Migration wird nie geändert.
- Regeln, die die Datenbank prüfen kann, gehören als Constraint hinein (Status, erledigt_am).

## Doppelte Stellen

Arten (`kategorien`) und Stati stehen mit ihren Namen im Server (`einstellungen.go`, für Export und
PDF) und in der Oberfläche (`frontend/src/lib/kategorien.js`, `status.js`); `einstellungen_test.go`
prüft, dass sie gleich heißen. Die Stati stehen zusätzlich im CHECK der Migration, die Arten
zusätzlich mit ihren Farbrollen `kat-<key>` in `rollen.css` (hell und dunkel). Eine Änderung braucht
alle Stellen. Welche Arten das Formular zeigt, schaltet die App selbst (Tabelle `arten_aus`).

## Datenschutz

Keine fremden Dienste, CDNs oder Schriften von außen. Einzige Ausnahme (Entscheidung vom 4. November
2026): Mitteilungen aufs Handy laufen über den Push-Dienst des Browsers, nur an die Dienste in
`pushDienste` (push.go) und Ende-zu-Ende verschlüsselt (webpush.go). Keine Namen oder Freitexte ins Log.
Neue Felder mit Personenbezug brauchen eine Löschfrist (`aufraeumen.go`) und eine Zeile auf der
Datenschutzseite.

## Git

Auf GitHub (`uuuxy/Schadenmeldung`, öffentlich, EUPL-1.2), Zweig `main`: Zugänge, Adressen und Pfade
des Servers stehen nur in `SERVER.lokal.md` (von Git ausgenommen), nie im Repo. Commit-Nachrichten auf Deutsch, ohne
Werkzeug-Hinweis, wie in der Bibliothek.

## Prüfen

- `docker compose -f docker-compose.local.yml up -d`, dann
  `TEST_DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' go test ./...`
- `go vet ./...` und `cd frontend && npm run build` (0 Warnungen); Skripte mit `shellcheck *.sh`
- Die schnellen Prüfungen laufen vor jedem Commit von selbst, wenn der Hook eingeschaltet ist:
  einmal je Klon `chmod +x .githooks/pre-commit && git config core.hooksPath .githooks`.
- Code auf Deutsch benennen wie in der Bibliothek; Kommentare erklären in ein, zwei Sätzen das Warum.
