// Die Stati einer Meldung mit Farbrolle: Jeder Stand, der in der Übersicht als Chip steht, hat seine
// eigene, damit der Hausmeister sie auf einen Blick auseinanderhält (StatusChip). Die Schlüssel prüfen auch der Server (server.go, stati) und
// die Datenbank (CHECK in migrations/002).
// „In Arbeit“ heißt: Der Hausmeister kümmert sich – ohne Zusage, wann. Das nimmt den Druck raus.
// `knapp` steht auf den Knöpfen des Hausmeisters, die in eine Handyzeile passen müssen.
export const stati = {
	offen: { name: 'Offen', knapp: 'Offen', icon: 'schedule', ton: 'neutral' },
	in_arbeit: { name: 'In Arbeit', knapp: 'In Arbeit', icon: 'handyman', ton: 'primaer' },
	bestellt: { name: 'Material bestellt', knapp: 'Bestellt', icon: 'local_shipping', ton: 'warten' },
	fachfirma: { name: 'Fachfirma / Kreis informiert', knapp: 'Fachfirma', icon: 'engineering', ton: 'extern' },
	erledigt: { name: 'Erledigt', knapp: 'Erledigt', icon: 'check_circle', ton: 'erfolg' }
};

// Was der Hausmeister zwischen „offen“ und „erledigt“ antippen kann.
export const zwischenstati = ['in_arbeit', 'bestellt', 'fachfirma'];

// Vermerke zum Antippen, damit niemand mit kalten Fingern im Flur tippen muss. Ohne Zeitzusage: Der
// Hausmeister soll sich nicht festlegen müssen.
export const schnellvermerke = [
	'Ersatzteil ist bestellt.',
	'Fachfirma ist beauftragt.',
	'Wartet auf den Schulträger.',
	'War nur ausgehängt – geht wieder.'
];
