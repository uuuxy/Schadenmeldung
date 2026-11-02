-- Neue Rolle verwaltung (Sekretariat): sieht die Übersicht und druckt QR-Aufkleber, ändert nichts.
-- Hausmeister, Leitung und Admin dürfen alles; das regelt der Server, nicht die Datenbank.
ALTER TABLE rollen DROP CONSTRAINT rollen_rolle_check;
ALTER TABLE rollen ADD CONSTRAINT rollen_rolle_check
	CHECK (rolle IN ('hausmeister', 'leitung', 'verwaltung', 'admin'));
