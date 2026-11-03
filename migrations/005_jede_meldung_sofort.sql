-- Keine Sammelmail mehr (Entscheidung vom 3. November 2026): Jede neue Meldung geht sofort als eigene
-- Mail hinaus. Je Person bleibt ein Schalter. Wer bisher überhaupt Mail wollte, bekommt sie weiter;
-- wer beides abbestellt hatte, bleibt ohne.
ALTER TABLE benachrichtigungen ADD COLUMN mail BOOLEAN;
UPDATE benachrichtigungen SET mail = dringend_sofort OR sammelmail;
ALTER TABLE benachrichtigungen
	ALTER COLUMN mail SET NOT NULL,
	DROP COLUMN dringend_sofort,
	DROP COLUMN sammelmail;

-- mail_versand ist jetzt der Postausgang: Die Zeile entsteht mit der Meldung, gesendet_am erst nach
-- dem Versand. Was nicht hinausging, schickt der Server später nach, ohne etwas doppelt zu senden.
ALTER TABLE mail_versand ALTER COLUMN gesendet_am DROP NOT NULL;

-- Der einzige Merkzettel darin war die Zeit der letzten Sammelmail.
DROP TABLE einstellungen;
