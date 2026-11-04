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

const datum = new Intl.DateTimeFormat('de-DE', { dateStyle: 'medium', timeStyle: 'short' });

/** Das Foto einer Meldung als Datei, fürs Teilen-Menü; ohne Netz geht es eben ohne Foto. */
export async function fotoAlsDatei(adresse) {
	try {
		const blob = await (await fetch(adresse)).blob();
		return new File([blob], 'schaden.jpg', { type: blob.type || 'image/jpeg' });
	} catch {
		return null;
	}
}

/**
 * Weiterleiten, etwa an den Schulträger oder eine Fachfirma, über das Teilen-Menü des Geräts: Text und
 * Foto gehen in die Mail-App oder wohin man will, Antworten landen im eigenen Postfach. Der Name der
 * meldenden Person bleibt draußen. Ohne Teilen-Menü (manche Rechner) öffnet sich eine Mail ohne Foto.
 */
export async function weiterleiten(meldung, foto) {
	const betreff = `${meldung.dringend ? 'Dringend' : 'Schaden'}: ${meldung.was} – ${meldung.ort}`;
	const text = [
		meldung.was + (meldung.dringend ? ' (dringend)' : ''),
		`Ort: ${meldung.ort}${meldung.bereich ? `, ${meldung.bereich}` : ''}`,
		meldung.notiz && `Notiz: ${meldung.notiz}`,
		`Gemeldet am ${datum.format(new Date(meldung.gemeldetAm))}`,
		zustand.einstellungen.schulname
	]
		.filter(Boolean)
		.join('\n');
	const daten = { title: betreff, text };
	if (foto && navigator.canShare?.({ files: [foto] })) daten.files = [foto];
	if (navigator.share) {
		try {
			await navigator.share(daten);
			return;
		} catch (e) {
			if (e.name === 'AbortError') return; // selbst abgebrochen
		}
	}
	location.href = `mailto:?subject=${encodeURIComponent(betreff)}&body=${encodeURIComponent(text)}`;
}
