// Die Stati einer Meldung mit Farbrolle. Die Schlüssel prüfen auch der Server (server.go, stati) und
// die Datenbank (CHECK in migrations/002).
// „Gesehen“ heißt: Der Hausmeister weiß Bescheid – ohne Zusage, wann. Das nimmt den Druck raus.
// `knapp` steht auf den Knöpfen des Hausmeisters, die in eine Handyzeile passen müssen.
export const stati = {
	offen: { name: 'Offen', knapp: 'Offen', icon: 'schedule', ton: 'neutral' },
	gesehen: { name: 'Gesehen', knapp: 'Gesehen', icon: 'visibility', ton: 'primaer' },
	bestellt: { name: 'Material bestellt', knapp: 'Bestellt', icon: 'local_shipping', ton: 'warten' },
	fachfirma: { name: 'Fachfirma / Kreis informiert', knapp: 'Fachfirma', icon: 'engineering', ton: 'warten' },
	erledigt: { name: 'Erledigt', knapp: 'Erledigt', icon: 'check_circle', ton: 'erfolg' }
};

// Was der Hausmeister zwischen „offen“ und „erledigt“ antippen kann.
export const zwischenstati = ['gesehen', 'bestellt', 'fachfirma'];

// Antworten zum Antippen, damit niemand mit kalten Fingern im Flur tippen muss.
export const schnellantworten = [
	'War nur ausgehängt – geht wieder.',
	'Bitte genauer beschreiben, wo.',
	'Ersatzteil ist bestellt.',
	'Kümmere mich heute darum.'
];
