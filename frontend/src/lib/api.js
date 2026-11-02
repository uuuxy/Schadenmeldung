export class ApiFehler extends Error {
	constructor(nachricht, status) {
		super(nachricht);
		this.status = status;
	}
}

/**
 * Ruft die API auf. Jede Anfrage trägt den Kopf X-Schaden, den der Server bei schreibenden Anfragen
 * als Schutz gegen Formulare fremder Seiten verlangt. Eine abgelaufene Sitzung meldet das Ereignis
 * „schaden:abgemeldet“, dann zeigt die App wieder die Anmeldung.
 */
export async function api(pfad, { methode = 'GET', json, formular } = {}) {
	const kopf = { 'X-Schaden': '1' };
	let body;
	if (json !== undefined) {
		kopf['Content-Type'] = 'application/json';
		body = JSON.stringify(json);
	} else if (formular) {
		body = formular;
	}

	let antwort;
	try {
		antwort = await fetch(pfad, { method: methode, headers: kopf, body, credentials: 'same-origin' });
	} catch {
		throw new ApiFehler('Keine Verbindung. Bitte WLAN oder mobile Daten prüfen.', 0);
	}

	const daten = antwort.headers.get('content-type')?.includes('application/json') ? await antwort.json() : null;
	if (antwort.status === 401 && pfad !== '/api/anmelden') {
		window.dispatchEvent(new Event('schaden:abgemeldet'));
	}
	if (!antwort.ok) {
		throw new ApiFehler(daten?.fehler ?? 'Das hat nicht geklappt. Bitte noch einmal versuchen.', antwort.status);
	}
	return daten;
}
