-- Kennung vom Handy: Schickt es dieselbe Meldung noch einmal (Antwort im Funkloch verloren, zwei
-- offene Tabs), legt der Server sie nur einmal an. Eindeutig je meldender Person; eine Zufallszahl
-- ohne Personenbezug, sie verschwindet mit der Meldung.
ALTER TABLE meldungen ADD COLUMN kennung TEXT;
CREATE UNIQUE INDEX meldungen_kennung ON meldungen (gemeldet_von, kennung);
