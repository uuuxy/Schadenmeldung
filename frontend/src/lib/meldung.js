import { api } from './api.js';
import { zustand, setzeStatus, uebernimm } from './zustand.svelte.js';
import { zeige } from './snackbar.svelte.js';
import { frage } from './rueckfrage.svelte.js';
import { rueckmeldung } from './haptik.js';

// Was man mit einer Meldung tun kann, an einer Stelle für Liste und Detailseite.

/** Sofort abhaken, ohne auf den Server zu warten; schlägt es fehl, kommt die Meldung zurück. */
export function abhaken(meldung) {
	const vorher = { ...meldung };
	uebernimm({ ...vorher, status: 'erledigt', erledigtAm: new Date().toISOString() });
	rueckmeldung();
	return setzeStatus(vorher.id, { status: 'erledigt' })
		.then(() =>
			zeige(`Erledigt: ${vorher.ort}`, {
				name: 'Rückgängig',
				tun: () => setzeStatus(vorher.id, { status: vorher.status }).catch((e) => zeige(e.message))
			})
		)
		.catch((e) => {
			uebernimm(vorher);
			zeige(e.message);
		});
}

/** Ein zweiter Tipp auf den gewählten Zwischenstand nimmt ihn zurück. */
export function zwischenstand(meldung, status) {
	return setzeStatus(meldung.id, { status: meldung.status === status ? 'offen' : status }).catch((e) => zeige(e.message));
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
