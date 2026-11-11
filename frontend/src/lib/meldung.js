import { api } from './api.js';
import { zustand, setzeStatus, uebernimm, mitAenderung } from './zustand.svelte.js';
import { legeStandInAusgang, nimmAusAusgang } from './nachreichen.js';
import { zeige } from './snackbar.svelte.js';
import { frage } from './rueckfrage.svelte.js';
import { rueckmeldung } from './haptik.js';
import { stati } from './status.js';

// Was man mit einer Meldung tun kann, an einer Stelle für Liste und Detailseite.

// Kein Netz, abgemeldet, gebremst oder Server gestört: Dann wartet die Änderung im Ausgang wie eine
// Meldung und geht raus, sobald es wieder geht (nachreichen.js). Sonst hat der Server abgelehnt.
const wartet = (e) => e.status === 0 || e.status === 401 || e.status === 429 || e.status >= 500;

/** Speichert eine Änderung am Stand; geht das gerade nicht, kommt sie in den Ausgang, und zurück kommt ihre Nummer dort. */
async function speichere(id, aenderung) {
	try {
		await setzeStatus(id, aenderung);
		return null;
	} catch (e) {
		if (!wartet(e)) throw e;
		if (e.status === 0) zustand.offline = true;
		return legeStandInAusgang(id, aenderung);
	}
}

/**
 * Setzt den Stand sofort in der Liste, ohne auf den Server zu warten. Im Funkloch bleibt er stehen und
 * geht raus, sobald wieder Netz da ist, denn der Hausmeister steht oft im Keller; lehnt der Server ab,
 * kommt der alte zurück. Die Snackbar bestätigt mit „Rückgängig“, denn wer hier tippt, ist schon
 * wieder in der Liste. aenderung kann einen ungespeicherten Vermerk mitbringen ({ antwort }).
 */
function sofort(meldung, aenderung, text) {
	const vorher = { ...meldung };
	uebernimm(mitAenderung(vorher, aenderung));
	rueckmeldung();
	return speichere(vorher.id, aenderung)
		.then((wartend) =>
			zeige(`${text}: ${vorher.ort}${wartend ? ' · wird nachgereicht' : ''}`, {
				name: 'Rückgängig',
				tun: () => zuruecknehmen(vorher, wartend)
			})
		)
		.catch((e) => {
			uebernimm(vorher);
			zeige(e.message);
		});
}

/** „Rückgängig“: Wartet die Änderung noch im Ausgang, fällt sie dort einfach weg; sonst geht der alte Stand hinterher. */
async function zuruecknehmen(vorher, wartend) {
	uebernimm(vorher);
	if (wartend && (await nimmAusAusgang(wartend))) return;
	await speichere(vorher.id, { status: vorher.status }).catch((e) => zeige(e.message));
}

export function abhaken(meldung, vermerk = {}) {
	return sofort(meldung, { status: 'erledigt', ...vermerk }, 'Erledigt');
}

/** Ein zweiter Tipp auf den gewählten Zwischenstand nimmt ihn zurück. */
export function zwischenstand(meldung, status, vermerk = {}) {
	const neu = meldung.status === status ? 'offen' : status;
	return sofort(meldung, { status: neu, ...vermerk }, neu === 'offen' ? 'Wieder offen' : stati[neu].name);
}

export function wiederOeffnen(meldung) {
	return sofort(meldung, { status: 'offen' }, 'Wieder offen');
}

/**
 * Dringend an oder aus, bei jeder Meldung: gilt sofort, die Übersicht sortiert gleich um. Wie ein Stand
 * wartet es im Funkloch im Ausgang; lehnt der Server ab, kommt der alte Wert zurück. Der Schalter zeigt
 * selbst, was gilt, eine Snackbar braucht es nur fürs Nachreichen oder bei einem Fehler.
 */
export async function dringendSetzen(meldung, dringend) {
	const vorher = { ...meldung };
	uebernimm(mitAenderung(vorher, { dringend }));
	rueckmeldung();
	try {
		if (await speichere(vorher.id, { dringend })) zeige(`${dringend ? 'Dringend' : 'Nicht mehr dringend'} · wird nachgereicht`);
	} catch (e) {
		uebernimm(vorher);
		zeige(e.message);
	}
}

export async function vermerkSpeichern(meldung, text) {
	const antwort = text.trim();
	try {
		const wartend = await speichere(meldung.id, { antwort });
		if (wartend) uebernimm({ ...meldung, antwort });
		zeige(`${antwort ? 'Vermerk gespeichert' : 'Vermerk entfernt'}${wartend ? ' · wird nachgereicht' : ''}`);
		return true;
	} catch (e) {
		zeige(e.message);
		return false;
	}
}

/** Löschen ist endgültig, auch das Foto; deshalb fragt die App einmal nach. */
export async function loeschen(meldung) {
	const ja = await frage({ titel: 'Meldung löschen?', text: `„${meldung.was}“ (${meldung.ort}) wird endgültig gelöscht, samt Foto.`, ja: 'Löschen' });
	if (!ja) return false;
	try {
		await api(`/api/meldungen/${meldung.id}`, { methode: 'DELETE' });
		zustand.offen = zustand.offen.filter((m) => m.id !== meldung.id);
		zustand.erledigt = zustand.erledigt.filter((m) => m.id !== meldung.id);
		zeige('Meldung gelöscht');
		return true;
	} catch (e) {
		zeige(e.message);
		return false;
	}
}
