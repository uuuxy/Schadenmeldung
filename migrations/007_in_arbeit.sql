-- „Gesehen“ heißt wieder „In Arbeit“ (Nutzer am 4. November): Wer es gesehen hat, kümmert sich auch;
-- zwei Namen für denselben Stand waren doppelt gemoppelt. Einen Zeitpunkt zeigt die App weiterhin nicht.
ALTER TABLE meldungen DROP CONSTRAINT meldungen_status_check;
UPDATE meldungen SET status = 'in_arbeit' WHERE status = 'gesehen';
ALTER TABLE meldungen ADD CONSTRAINT meldungen_status_check
	CHECK (status IN ('offen', 'in_arbeit', 'bestellt', 'fachfirma', 'erledigt'));
