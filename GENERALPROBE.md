# Generalprobe vor dem Start an der Schule

Einmal auf dem Schulserver durchspielen, bevor das Kollegium die App bekommt. Dauer etwa eine Stunde,
mit zwei Handys (ein iPhone, ein Android-Handy), dem Hausmeister und dem Sekretariat. Jeder Schritt nennt,
was passieren muss. Passiert etwas anderes, ist die Probe nicht bestanden.

## 1. Vorher einrichten

- [ ] **Ausnahme beim Mail-Server:** Die Schul-IT trägt die Adresse des Servers als Ausnahme ein
      (fail2ban, `ignoreip`). Ohne sie sperren ein paar falsche Passwörter alle aus, bei der Anmeldung
      und beim Mailversand.
- [ ] **`.env`:** `SICHERUNG_KOPIE` (zweites Gerät, etwa das NAS) und `WAECHTER_AN` (wer Warnungen
      bekommt) eintragen, siehe `.env.example`.
- [ ] **Cron:** Wächter und nächtliche Sicherung eintragen, siehe README, „Betrieb auf dem Server“.
- [ ] **`./sicherung.sh` einmal von Hand:** Es meldet „Gesichert …“ und „Kopiert nach …“, und die
      Datei liegt auf dem zweiten Gerät.
- [ ] **`./sicherung-pruefen.sh`:** „✓ Die Sicherung lässt sich zurückspielen“.
- [ ] **`./waechter.sh --probe`:** Die Probemail kommt an.

## 2. Mit echten Handys

Lehrkraft A mit iPhone, Lehrkraft B mit Android-Handy, der Hausmeister mit seinem Handy.

- [ ] **Installieren (A und B):** Die Karte „Als App …“ erscheint gleich beim ersten Öffnen. Den
      Schritten folgen: Das Symbol liegt auf dem Startbildschirm. A meldet sich in der App neu an.
- [ ] **Melden mit QR-Code (A):** Aufkleber scannen, der Raum ist gewählt, Foto machen, abschicken:
      „Gemeldet – danke!“. Der Hausmeister hat sofort eine Mail, mit eingeschalteten Mitteilungen auch
      eine auf dem Handy.
- [ ] **Melden ohne QR-Code (B):** Den Raum über die Kacheln wählen und melden.
- [ ] **Doppelt (B):** Den Raum von A wählen: Unter „Hier schon gemeldet“ steht die Meldung von A.
- [ ] **Funkloch beim Melden (B):** Flugmodus an und WLAN aus, melden: „Wartet auf Netz“. Wieder
      Netz: „Nachgereicht“, und die Meldung kommt genau einmal an.
- [ ] **Knopf in der Mail (Hausmeister):** In der Mail auf „In Arbeit“ tippen: Eine Seite öffnet sich,
      geändert ist noch nichts. Erst der Knopf dort setzt den Stand, und die Übersicht zeigt „In Arbeit“.
- [ ] **In der App (Hausmeister):** Eine Meldung öffnen und „Bestellt“ tippen: Die App geht zurück zur
      Übersicht. „Rückgängig“ in der Snackbar nimmt es zurück.
- [ ] **Funkloch beim Hausmeister:** Flugmodus an und WLAN aus, eine Meldung abhaken: Die Snackbar
      sagt „wird nachgereicht“, unter „Erledigt“ steht „wartet auf Netz“. Wieder Netz: „Änderung
      nachgereicht“, und auf einem zweiten Gerät erscheint der Stand ohne Neuladen.
- [ ] **Mail an die Meldenden (A):** 10 bis 15 Minuten nach „Erledigt“ hat A die Mail „Erledigt –
      danke fürs Melden!“, ohne den internen Vermerk.
- [ ] **Sekretariat:** Sieht die Übersicht und druckt einen QR-Bogen, kann aber nichts ändern.
- [ ] **Schulleitung:** Sieht neue Meldungen ohne Neuladen; der Export öffnet sich.

## 3. Ausfall proben

- [ ] **App hängt:** `docker compose stop schaden`. Nach 5 bis 10 Minuten läuft sie von selbst wieder;
      `journalctl -t schadenmeldung` zeigt „Starte neu: schaden“.
- [ ] **Server startet neu:** `reboot`. Danach laufen App und Datenbank von selbst, alle Meldungen
      sind da.
- [ ] **Sicherung zurückspielen:** `./sicherung-pruefen.sh` zeigt so viele Meldungen, wie die App
      unter „Offen“ und „Erledigt“ hat.

## 4. Danach

- [ ] Die Testmeldungen löschen (Seite der Meldung, „Löschen“).
- [ ] QR-Aufkleber kleben; Hausmeister und Sekretariat in den Einstellungen eintragen, falls noch
      nicht geschehen.
- [ ] Das Kollegium einweisen, am besten in einer Konferenz: Die Karte führt durchs Installieren, die
      Anmeldung läuft mit der Schul-Mail. Die Ausnahme beim Mail-Server muss vorher stehen, denn dort
      tippen viele gleichzeitig ihr Passwort.
