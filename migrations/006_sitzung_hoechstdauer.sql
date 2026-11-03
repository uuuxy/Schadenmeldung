-- Eine Sitzung verlängert sich bei Nutzung, aber höchstens bis ein Jahr nach der Anmeldung
-- (sitzungen.go). Bestehende Sitzungen zählen ab jetzt.
ALTER TABLE sitzungen ADD COLUMN angelegt_am TIMESTAMPTZ NOT NULL DEFAULT now();
