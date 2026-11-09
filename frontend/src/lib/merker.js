// Reiter und Filter der Übersicht, nur kurz gemerkt, damit „Zurück“ aus einer Meldung dort landet. Wer
// die App später wieder öffnet, beginnt bei „Offen“ ohne Filter, sonst hielte man die erledigten
// Meldungen von gestern für die offenen. Die Seite einer Meldung setzt den Reiter, wenn sie ihn ändert.
const SCHLUESSEL = 'schaden.uebersicht';
export const MERKZEIT = 10 * 60 * 1000;

export function gemerkt() {
	try {
		const m = JSON.parse(sessionStorage.getItem(SCHLUESSEL) ?? '{}');
		return Date.now() - (m.zeit ?? 0) < MERKZEIT ? m : {};
	} catch {
		return {};
	}
}

export function merke(stand) {
	try {
		sessionStorage.setItem(SCHLUESSEL, JSON.stringify({ ...stand, zeit: Date.now() }));
	} catch {
		/* Ohne Speicher beginnt die Übersicht eben bei „Offen“. */
	}
}
