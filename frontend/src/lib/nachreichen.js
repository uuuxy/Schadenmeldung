import { api } from './api.js';
import { legeAb, entferne, nimmHeraus } from './ausgang.js';
import { zustand, uebernimm, setzeStatus, ausgangVon, ladeAusgang, ladeMeldungen, ladeEinstellungen, merkeStand } from './zustand.svelte.js';
import { zeige } from './snackbar.svelte.js';

// Nachreichen: Was im Funkloch nicht rausging, Meldungen und Änderungen am Stand, wartet im Ausgang
// (ausgang.js) und geht der Reihe nach raus, sobald wieder Netz da ist. App.svelte frischt dafür beim
// Start, alle 30 Sekunden, beim Sichtbarwerden und bei zurückkehrendem Netz auf. Jeder Eintrag merkt sich,
// wer ihn gespeichert hat (von): Endet die Sitzung und meldet sich danach eine andere Person am Gerät
// an, geht er nicht unter deren Namen raus, sondern wartet auf seine Person.

/** Schickt Wartendes los und holt den neuesten Stand; im Funkloch bleibt der letzte Stand stehen. */
export async function auffrischen() {
	if (!zustand.benutzer) return;
	try {
		await sendeAusgang();
		// Die Einstellungen kommen mit: Schaltet jemand eine Art ab, verschwindet ihre Kachel bald überall.
		await Promise.all([ladeMeldungen(), ladeEinstellungen()]);
		merkeStand();
	} catch (e) {
		if (e.status === 0) zustand.offline = true;
	}
}

/** Legt eine Meldung im Ausgang ab, wenn sie gerade nicht rausgeht; zurück kommt ihre Nummer dort. */
export async function legeInAusgang(felder, foto, anzeige) {
	// Ihr Zeitpunkt reist mit.
	const id = await legeAb({ von: zustand.benutzer?.email, felder, foto, anzeige, erstellt: new Date().toISOString() });
	await ladeAusgang();
	return id;
}

/**
 * Legt eine Änderung am Stand in den Ausgang, wenn sie gerade nicht rausgeht; zurück kommt ihre Nummer dort.
 * vorher reist mit: Erledigt jemand anderes die Meldung, während die Änderung wartet, öffnet sie sie nicht wieder.
 */
export async function legeStandInAusgang(meldungId, aenderung, vorher) {
	const id = await legeAb({ von: zustand.benutzer?.email, art: 'stand', meldungId, aenderung, vorher, erstellt: new Date().toISOString() });
	await ladeAusgang();
	return id;
}

/** Für „Rückgängig“: true, wenn die Änderung noch wartete und jetzt nicht mehr rausgeht. */
export async function nimmAusAusgang(id) {
	const eintrag = await nimmHeraus(id).catch(() => undefined);
	await ladeAusgang();
	return Boolean(eintrag);
}

// Läuft schon ein Senden, wartet ein weiterer Aufruf darauf, statt gleich zurückzukehren. Sonst sähe
// „Abmelden“ als wartend, was gerade unterwegs ist, und fragte grundlos nach.
let laufend = null;

function sendeAusgang() {
	if (!zustand.benutzer) return;
	laufend ??= sendeDerReiheNach().finally(() => (laufend = null));
	return laufend;
}

async function sendeDerReiheNach() {
	let staende = 0;
	try {
		// Der Reihe nach, wie getippt: Zwei Änderungen an derselben Meldung kommen so in ihrer Folge an.
		for (const eintrag of await ausgangVon(zustand.benutzer?.email)) {
			const stand = eintrag.art === 'stand';
			try {
				if (stand) {
					await setzeStatus(eintrag.meldungId, eintrag.aenderung, eintrag.vorher);
					staende++;
				} else {
					const formular = new FormData();
					for (const [schluessel, wert] of Object.entries(eintrag.felder)) formular.set(schluessel, wert);
					formular.set('erstellt', eintrag.erstellt);
					if (eintrag.foto) formular.set('foto', eintrag.foto, 'foto.jpg');
					uebernimm(await api('/api/meldungen', { methode: 'POST', formular }));
					zustand.nachgereicht.push(eintrag.id);
					zeige(`Nachgereicht: ${eintrag.anzeige.was} (${eintrag.anzeige.ort})`);
				}
				await entferne(eintrag.id);
				zustand.offline = false;
			} catch (e) {
				// Kein Netz, abgemeldet, gebremst oder Server gestört: liegen lassen und später erneut versuchen.
				if (e.status === 0 || e.status === 401 || e.status === 429 || e.status >= 500) throw e;
				// Abgelehnt (etwa weil Raum oder Meldung inzwischen gelöscht sind): nicht endlos wiederholen.
				await entferne(eintrag.id);
				zeige(stand ? `Eine Änderung wurde abgelehnt: ${e.message}` : `Eine gespeicherte Meldung wurde abgelehnt: ${e.message}`);
			}
		}
	} finally {
		await ladeAusgang();
		if (staende) zeige(staende === 1 ? 'Änderung nachgereicht' : `${staende} Änderungen nachgereicht`);
	}
}
