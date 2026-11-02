// Die Druckformate der QR-Aufkleber, alle auf A4 (210 × 297 mm). Maße in mm, wie sie auf der
// Packung stehen; Etikettenbögen füllen das Blatt lückenlos, deshalb genügen der Rand oben und
// links. Schrift in pt, denn sie landet auf Papier. „seite“ stellt den Code neben den Text (für
// breite Etiketten), „stapel“ darüber. Die Packungsnummern sind Beispiele: Jeder Bogen derselben
// Maße passt.
export const formate = [
	{
		key: 'schnitt',
		name: 'Zum Ausschneiden',
		kurz: '12 je Blatt',
		hinweis: 'Normales Papier, mit Schnittlinien zum Ausschneiden.',
		spalten: 3,
		zeilen: 4,
		breite: 63.33,
		hoehe: 69.25,
		oben: 10,
		links: 10,
		linien: true,
		anordnung: 'stapel',
		qr: 36,
		schrift: { marke: 8, name: 15, bereich: 9, tipp: 7.5 }
	},
	{
		key: 'gross',
		name: 'Etiketten groß',
		kurz: '8 je Blatt',
		hinweis: 'Etikettenbogen 105 × 74 mm, 8 je A4-Blatt, z. B. Avery Zweckform 3427.',
		spalten: 2,
		zeilen: 4,
		breite: 105,
		hoehe: 74,
		oben: 0.5,
		links: 0,
		anordnung: 'seite',
		qr: 48,
		schrift: { marke: 9, name: 24, bereich: 11, tipp: 8 }
	},
	{
		key: 'klein',
		name: 'Etiketten klein',
		kurz: '24 je Blatt',
		hinweis: 'Etikettenbogen 70 × 37 mm, 24 je A4-Blatt, z. B. Avery Zweckform 3474.',
		spalten: 3,
		zeilen: 8,
		breite: 70,
		hoehe: 37,
		oben: 0.5,
		links: 0,
		anordnung: 'seite',
		qr: 27,
		schrift: { marke: 6.5, name: 14, bereich: 7.5, tipp: 0 }
	},
	{
		key: 'blatt',
		name: 'Einzelblatt',
		kurz: '1 je Blatt',
		hinweis: 'Ein Raum je A4-Blatt, groß für Tür oder Aushang. Normales Papier.',
		spalten: 1,
		zeilen: 1,
		breite: 210,
		hoehe: 297,
		oben: 0,
		links: 0,
		anordnung: 'stapel',
		qr: 120,
		schrift: { marke: 16, name: 40, bereich: 18, tipp: 14 }
	}
];

export const formatNach = (key) => formate.find((f) => f.key === key) ?? formate[0];

/** Teilt die Räume auf Blätter: so viele je Blatt, wie das Format Felder hat. */
export function aufBlaetter(raeume, format) {
	const jeBlatt = format.spalten * format.zeilen;
	const blaetter = [];
	for (let i = 0; i < raeume.length; i += jeBlatt) blaetter.push(raeume.slice(i, i + jeBlatt));
	return blaetter;
}

// Das gewählte Format merkt sich das Gerät; ohne Speicher beginnt es beim ersten.
const SCHLUESSEL = 'schaden.qrFormat';

export function gemerktesFormat() {
	try {
		return localStorage.getItem(SCHLUESSEL) ?? formate[0].key;
	} catch {
		return formate[0].key;
	}
}

export function merkeFormat(key) {
	try {
		localStorage.setItem(SCHLUESSEL, key);
	} catch {
		/* Ohne Speicher eben jedes Mal das erste Format. */
	}
}
