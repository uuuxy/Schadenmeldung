# Arbeitsregeln für dieses Projekt

Die Schadensmeldung soll schlank bleiben: ein Go-Programm mit eingebetteter Svelte-Oberfläche und
PostgreSQL, keine weiteren Dienste. Bevor etwas Neues dazukommt: Braucht es das wirklich?

Sie heißt „Schadensmeldung“, mit Fugen-s, wie man es meist sagt und schreibt (Entscheidung vom
10. November 2026). Technische Namen bleiben „schadenmeldung“: GitHub, Docker-Image, Server und der Ausgang im
Browser (`ausgang.js`). Das sieht niemand, und eine Umbenennung zöge den Server mit oder ließe
wartende Meldungen verwaisen.

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
- **M3 Expressive gezielt** (Entscheidung vom 5. November 2026): Knöpfe und Kacheln werden beim Drücken
  eckiger und bleiben es kurz (Shape morph), Fragen steigen beim Aufklappen federnd an ihren Platz
  (`lib/bewegung.js`, mit Googles Kurven fürs Web aus den Motion specs), das Häkchen bei „Gemeldet“ steht in einer Expressive-Form. Nur Räumliches federt, Farbe,
  Deckkraft und Zuklappen nicht; „Bewegung reduzieren“ schaltet alles ab. Nicht übernehmen: schwebende
  Werkzeugleisten, Knopfgruppen, wellige Fortschrittsbalken, verspielte Formen in der Übersicht.
- **Zwei Ebenen:** Wer die Übersicht sieht (Hausmeister, Leitung, Verwaltung), wechselt in der
  Kopfleiste per Segmented Button zwischen „Übersicht“ und „Melden“ (`KopfLeiste ebene=…`), nicht
  per Switch: M3 nimmt Switches nur für Ein/Aus. Die Übersicht aktualisiert sich live über
  `/api/ereignisse`; jeder neue Schreibweg für Meldungen ruft `a.verteiler.melde()`.
- **Liste und Detail (M3 list-detail):** Eine Zeile der Übersicht trägt bewusst nur eine Zusatzaktion
  (das Häkchen „Erledigt“), damit die Liste ruhig bleibt; alles Weitere steht auf der Seite der Meldung
  (`/meldung/<id>`). M3 erlaubt mehrere Aktionen je Zeile (Stand November 2026), mehr als drei Knöpfe
  nebeneinander sind aber zu viel. Endgültiges fragt über `frage()` nach (`lib/rueckfrage.svelte.js`),
  nie über `confirm()`. Mehrere Meldungen wählt man nach M3 (selection, wie Gmail): lange drücken oder
  ⋮ → „Drucken“, dann ist die ganze Zeile die Auswahl, ohne „Erledigt“; die Kopfleiste zeigt Kreuz, Zahl
  und Aktionen (`AuswahlLeiste`). Gedruckt wird als PDF vom Server (`druck.go`). Texte nach M3 knapp:
  Menüeinträge nur Symbol und Wort, keine Erklärzeilen; was die Oberfläche zeigt, braucht keine Anleitung.
- **Kacheln:** In einem Raster sind alle Kacheln gleich groß (M3: das Standard-Raster zeigt Karten
  gleich groß), die letzte Reihe darf kürzer sein; Form `rounded-xl`, Fläche `bg-surface-container-low`,
  Abstand `gap-3`. Passt ein Name nicht, bekommt die Kachel ein Symbol statt einer größeren Fläche.
  Pillenform nur für einzeilige Knöpfe, Suche und Segmente (M3 „round“), Chips `rounded-sm`.
- **Zurück:** Jede Unteransicht (ein Bereich, „Weitere Räume“, „Anderer Ort“) hat oben den Pfeil mit
  ihrem Namen (`ZurueckZeile`), nie einen Textlink darunter (M3: der Zurück-Pfeil steht vorn). Die Zeile
  gibt ihr einen Eintrag im Verlauf (`unteransicht`, weg.svelte.js), so führen Pfeil, Zurück-Geste von
  Android und Wischen in Safari gleich weit; in der App vom Home-Bildschirm am iPhone wischt
  `ZurueckWischen` (tippt den innersten Pfeil mit `data-zurueck` an). Neue Unteransichten nehmen deshalb
  immer `ZurueckZeile`, neue Pfeile tragen `data-zurueck`.
- **Farben nur über Rollen** aus `frontend/src/styles/rollen.css` (`bg-surface`,
  `text-on-surface-variant`, `bg-primary-container` …), nie Palettenfarben. Jede Rolle hat einen
  Wert im hellen und im dunklen Schema. Einzige Ausnahme: die Papier-Rollen (`papier`, `auf-papier`,
  `papier-akzent` …) für den QR-Bogen bleiben fest, denn Papier ist weiß.
- **Schrift über die M3-Stufen** aus `mass.css` (`text-title-medium`, `text-body-large`,
  `text-label-large` …), Formen über die M3-Radien (`rounded-sm` = Chip, `rounded-xl` = Karte
  und Kachel, `rounded-2xl` = großes Symbolfeld/Blatt).
- **Symbole:** Material Symbols Rounded. Neues Symbol in `frontend/scripts/symbole.mjs` eintragen,
  dann `npm run symbole`.
- **Wenig tippen:** Neue Eingaben zuerst als Kacheln oder Chips denken, Freitext nur als Ausweg.
  Alles muss auch ohne QR-Code in wenigen Tipps gehen.
- **Funkloch:** Schreibende Wege dürfen im Funkloch nichts verlieren: Meldungen und Stand-Änderungen
  warten im Ausgang (`lib/ausgang.js`, `lib/nachreichen.js`), bis wieder Netz da ist, und
  `ladeMeldungen` legt wartende Änderungen über den Stand vom Server. Abmelden schickt zuerst los
  und fragt nach, statt Wartendes still zu löschen. Was wartet, geht nur unter der Person raus, die
  es gespeichert hat (`von`), auch wenn sich danach jemand anderes am Gerät anmeldet. Der Service
  Worker cacht nie `/api/` oder `/fotos/`.

## Rollen und Mail

- Kollegium sieht nur das Meldeformular; Hausmeister, Leitung und Admin dürfen alles, die
  Verwaltung (Sekretariat) sieht und druckt QR-Aufkleber, ändert aber nichts (`rollen.go`,
  `zustand.svelte.js`). Jede neue Tür prüft die Rolle im Server, nicht nur die Oberfläche.
- Das Postfach für den Versand steht fest in der .env (`MAIL_ABSENDER`) oder wird unter Einstellungen →
  Mail eingetragen, am besten von der Person, der es gehört: Testmail vor dem Speichern, das Passwort
  nur verschlüsselt (`SCHLUESSEL` in der .env, `update.sh` legt ihn an) und nie wieder angezeigt.
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
Der QR-Scanner (`QrScanner.svelte`) liest auf dem Handy selbst (jsQR, erst beim Öffnen geladen); das
Kamerabild geht nie an den Server.
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
