// Die zuletzt gewählten Räume dieses Geräts, als Abkürzung beim nächsten Melden. Nur im Browser
// gespeichert; fehlt der Speicher, gibt es eben keine Abkürzung.
const SCHLUESSEL = 'schaden.zuletzt';

export function zuletztGewaehlt() {
	try {
		return JSON.parse(localStorage.getItem(SCHLUESSEL) ?? '[]');
	} catch {
		return [];
	}
}

export function merkeRaum(id) {
	try {
		const liste = [id, ...zuletztGewaehlt().filter((x) => x !== id)].slice(0, 4);
		localStorage.setItem(SCHLUESSEL, JSON.stringify(liste));
	} catch {
		/* Ohne Speicher keine Abkürzung. */
	}
}
