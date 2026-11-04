import { api } from './api.js';
import { zustand, setzeStatus, uebernimm } from './zustand.svelte.js';
import { zeige } from './snackbar.svelte.js';
import { frage } from './rueckfrage.svelte.js';
import { rueckmeldung } from './haptik.js';
import { stati } from './status.js';

// Was man mit einer Meldung tun kann, an einer Stelle für Liste und Detailseite.

/**
 * Setzt den Stand sofort in der Liste, ohne auf den Server zu warten; schlägt es fehl, kommt der alte
 * zurück. Die Snackbar bestätigt mit „Rückgängig“, denn wer hier tippt, ist schon wieder in der Liste.
 * aenderung kann einen ungespeicherten Vermerk mitbringen ({ antwort }).
 */
function sofort(meldung, aenderung, text) {
	const vorher = { ...meldung };
	uebernimm({ ...vorher, ...aenderung, ...(aenderung.status === 'erledigt' && { erledigtAm: new Date().toISOString() }) });
	rueckmeldung();
	return setzeStatus(vorher.id, aenderung)
		.then(() =>
			zeige(`${text}: ${vorher.ort}`, {
				name: 'Rückgängig',
				tun: () => setzeStatus(vorher.id, { status: vorher.status }).catch((e) => zeige(e.message))
			})
		)
		.catch((e) => {
			uebernimm(vorher);
			zeige(e.message);
		});
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
	return setzeStatus(meldung.id, { status: 'offen' })
		.then(() => zeige('Wieder offen'))
		.catch((e) => zeige(e.message));
}

export async function vermerkSpeichern(meldung, text) {
	try {
		await setzeStatus(meldung.id, { antwort: text });
		zeige(text.trim() ? 'Vermerk gespeichert' : 'Vermerk entfernt');
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
