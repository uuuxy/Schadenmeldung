-- Arten, die das Meldeformular nicht anbietet. Ohne Zeile ist eine Art sichtbar; abschalten und
-- wieder anschalten lässt sie sich unter Einstellungen → Arten. „Sonstiges“ bleibt immer, als
-- Ausweg für alles, was in keine Kachel passt.
CREATE TABLE arten_aus (
	kategorie TEXT PRIMARY KEY CHECK (kategorie <> 'sonstiges')
);

-- Medientechnik (PC, Beamer, Display) meldet das Kollegium über das Ticketsystem der IT im
-- Landesportal; die Schadenmeldung soll ihm keine Konkurrenz machen. Deshalb zu Beginn aus.
INSERT INTO arten_aus (kategorie) VALUES ('technik');
