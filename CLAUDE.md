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
  Bibliothek (dort steht jeweils die M3-Begründung). Auswahl liegt immer auf secondary-container,
  Knöpfe tragen keinen Schatten, Chips brechen um statt seitlich zu scrollen.
- **Farben nur über Rollen** aus `frontend/src/styles/rollen.css` (`bg-surface`,
  `text-on-surface-variant`, `bg-primary-container` …), nie Palettenfarben. Jede Rolle hat einen
  Wert im hellen und im dunklen Schema. Einzige Ausnahme: der QR-Bogen, denn Papier ist weiß.
- **Schrift über die M3-Stufen** aus `mass.css` (`text-title-medium`, `text-body-large`,
  `text-label-large` …), Formen über die M3-Radien (`rounded-sm` = Chip, `rounded-xl` = Karte,
  `rounded-2xl` = Kachel/Blatt).
- **Symbole:** Material Symbols Rounded. Neues Symbol in `frontend/scripts/symbole.mjs` eintragen,
  dann `npm run symbole`.
- **Wenig tippen:** Neue Eingaben zuerst als Kacheln oder Chips denken, Freitext nur als Ausweg.
  Alles muss auch ohne QR-Code in wenigen Tipps gehen.
- **Funkloch:** Schreibende Wege der Kollegien-Seite dürfen im Funkloch nichts verlieren
  (Ausgang in `lib/ausgang.js`, `zustand.sendeAusgang`). Der Service Worker cacht nie `/api/` oder
  `/fotos/`.

## Rollen und Mail

- Kollegium sieht nur das Meldeformular; Hausmeister bearbeitet, Leitung liest, Verwaltung darf
  alles (`rollen.go`, `zustand.svelte.js`). Jede neue Tür prüft die Rolle im Server, nicht nur die
  Oberfläche.
- Mails sollen den Hausmeister nicht hetzen: keine Erinnerungen, keine Eskalation, keine Zeitstempel
  von „Gesehen“ in der Oberfläche. Mail-Links lösen nie beim Öffnen etwas aus, nur beim Antippen.
- Lokal landen Mails in Mailpit (`docker-compose.local.yml`, http://localhost:8025).

## Datenbank

- PostgreSQL mit pgx. Schemaänderungen nur als neue, nummerierte Datei in `migrations/`; eine
  eingespielte Migration wird nie geändert.
- Regeln, die die Datenbank prüfen kann, gehören als Constraint hinein (Status, erledigt_am).

## Doppelte Stellen

Arten (`kategorien`) und Stati stehen im Server (`server.go`) und in der Oberfläche
(`frontend/src/lib/kategorien.js`, `status.js`); die Stati zusätzlich im CHECK der Migration. Eine
Änderung braucht alle Stellen.

## Datenschutz

Keine fremden Dienste, CDNs oder Schriften von außen. Keine Namen oder Freitexte ins Log.
Neue Felder mit Personenbezug brauchen eine Löschfrist (`aufraeumen.go`) und eine Zeile auf der
Datenschutzseite.

## Git

Privat auf GitHub (`uuuxy/schaden`), Zweig `main`. Commit-Nachrichten auf Deutsch, ohne
Werkzeug-Hinweis, wie in der Bibliothek.

## Prüfen

- `docker compose -f docker-compose.local.yml up -d`, dann
  `TEST_DATABASE_URL='postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable' go test ./...`
- `go vet ./...` und `cd frontend && npm run build` (0 Warnungen)
- Code auf Deutsch benennen wie in der Bibliothek; Kommentare erklären in ein, zwei Sätzen das Warum.
