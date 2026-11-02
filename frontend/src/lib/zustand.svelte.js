import { api } from './api.js';
import { legeAb, alleImAusgang, entferne, leereAusgang } from './ausgang.js';
import { zeige } from './snackbar.svelte.js';

/** Der gemeinsame Zustand der App: wer angemeldet ist, die Meldungen, die Räume und das Netz. */
export const zustand = $state({
	bereit: false,
	benutzer: null,
	einstellungen: { schulname: '', domain: '', fotoTage: 30, aufbewahrungTage: 365 },
	offen: [],
	erledigt: [],
	raeume: [],
	// Funkloch: Die App zeigt den letzten Stand (standVon) und sammelt Meldungen im Ausgang.
	offline: false,
	standVon: null,
	ausgang: []
});

// Kollegium meldet nur. Hausmeister bearbeitet, Leitung sieht alles (ändert nichts), Verwaltung darf
// alles. Der Server prüft dasselbe noch einmal.
export const darfBearbeiten = () => ['hausmeister', 'admin'].includes(zustand.benutzer?.rolle);
export const darfSehen = () => ['hausmeister', 'leitung', 'admin'].includes(zustand.benutzer?.rolle);
export const istAdmin = () => zustand.benutzer?.rolle === 'admin';

// ── Letzter Stand auf dem Gerät ──────────────────────────────────────────────
// Damit die App im Funkloch nicht leer oder mit der Anmeldung startet. Beim Abmelden gelöscht.
const STAND = 'schaden.stand';

function merkeStand() {
	try {
		const { benutzer, einstellungen, raeume, offen, erledigt } = $state.snapshot(zustand);
		localStorage.setItem(STAND, JSON.stringify({ zeit: new Date().toISOString(), benutzer, einstellungen, raeume, offen, erledigt }));
	} catch {
		/* Speicher voll oder gesperrt: dann eben ohne Offline-Stand. */
	}
}

function gemerkterStand() {
	try {
		return JSON.parse(localStorage.getItem(STAND) ?? 'null');
	} catch {
		return null;
	}
}

/** Beim Abmelden: nichts von dieser Person bleibt auf dem Gerät. */
export async function vergissAlles() {
	try {
		localStorage.removeItem(STAND);
	} catch {
		/* nichts zu tun */
	}
	await leereAusgang().catch(() => {});
	zustand.ausgang = [];
	zustand.offen = [];
	zustand.erledigt = [];
}

// ── Laden ────────────────────────────────────────────────────────────────────

export async function starte() {
	try {
		zustand.einstellungen = await api('/api/einstellungen');
	} catch {
		/* Ohne Einstellungen geht es mit den Vorgaben weiter. */
	}
	try {
		zustand.benutzer = await api('/api/ich');
		await ladeAlles();
	} catch (e) {
		const stand = e.status === 0 ? gemerkterStand() : null;
		if (stand?.benutzer) {
			Object.assign(zustand, {
				benutzer: stand.benutzer,
				einstellungen: stand.einstellungen,
				raeume: stand.raeume,
				offen: stand.offen,
				erledigt: stand.erledigt,
				standVon: stand.zeit,
				offline: true
			});
		} else {
			zustand.benutzer = null;
		}
	}
	await ladeAusgang();
	zustand.bereit = true;
}

export async function ladeAlles() {
	await Promise.all([ladeMeldungen(), ladeRaeume()]);
}

export async function ladeMeldungen() {
	const daten = await api('/api/meldungen');
	zustand.offen = daten.offen;
	zustand.erledigt = daten.erledigt;
	zustand.offline = false;
	zustand.standVon = null;
	merkeStand();
}

export async function ladeRaeume() {
	zustand.raeume = await api('/api/raeume');
	merkeStand();
}

/** Schickt Wartendes los und holt den neuesten Stand; im Funkloch bleibt der letzte Stand stehen. */
export async function auffrischen() {
	if (!zustand.benutzer) return;
	try {
		await sendeAusgang();
		await ladeMeldungen();
	} catch (e) {
		if (e.status === 0) zustand.offline = true;
	}
}

// ── Meldungen ändern ─────────────────────────────────────────────────────────

// Dringendes zuerst, sonst das Neueste oben – dieselbe Ordnung wie auf dem Server.
function sortiereOffen(liste) {
	return liste.sort((a, b) => b.dringend - a.dringend || new Date(b.gemeldetAm) - new Date(a.gemeldetAm));
}

/** Übernimmt eine geänderte Meldung in die passende Liste, ohne alles neu zu laden. */
export function uebernimm(meldung) {
	const offen = zustand.offen.filter((m) => m.id !== meldung.id);
	const erledigt = zustand.erledigt.filter((m) => m.id !== meldung.id);
	if (meldung.status === 'erledigt') {
		zustand.offen = offen;
		zustand.erledigt = [meldung, ...erledigt];
	} else {
		zustand.offen = sortiereOffen([...offen, meldung]);
		zustand.erledigt = erledigt;
	}
	merkeStand();
}

export async function setzeStatus(id, aenderung) {
	const neu = await api(`/api/meldungen/${id}/status`, { methode: 'POST', json: aenderung });
	uebernimm(neu);
	return neu;
}

// ── Ausgang ──────────────────────────────────────────────────────────────────

async function ladeAusgang() {
	try {
		zustand.ausgang = (await alleImAusgang()).map(({ id, anzeige, erstellt }) => ({ id, ...anzeige, erstellt }));
	} catch {
		zustand.ausgang = [];
	}
}

/** Legt eine Meldung im Ausgang ab, wenn sie gerade nicht rausgeht. Ihr Zeitpunkt reist mit. */
export async function legeInAusgang(felder, foto, anzeige) {
	await legeAb({ felder, foto, anzeige, erstellt: new Date().toISOString() });
	await ladeAusgang();
}

let sendetGerade = false;

export async function sendeAusgang() {
	if (sendetGerade || !zustand.benutzer) return;
	sendetGerade = true;
	try {
		for (const eintrag of await alleImAusgang()) {
			const formular = new FormData();
			for (const [schluessel, wert] of Object.entries(eintrag.felder)) formular.set(schluessel, wert);
			formular.set('erstellt', eintrag.erstellt);
			if (eintrag.foto) formular.set('foto', eintrag.foto, 'foto.jpg');
			try {
				uebernimm(await api('/api/meldungen', { methode: 'POST', formular }));
				await entferne(eintrag.id);
				zustand.offline = false;
				zeige(`Nachgereicht: ${eintrag.anzeige.was} (${eintrag.anzeige.ort})`);
			} catch (e) {
				// Kein Netz, abgemeldet oder Server gestört: liegen lassen und später erneut versuchen.
				if (e.status === 0 || e.status === 401 || e.status >= 500) throw e;
				// Abgelehnt (etwa weil der Raum inzwischen gelöscht ist): nicht endlos wiederholen.
				await entferne(eintrag.id);
				zeige(`Eine gespeicherte Meldung wurde abgelehnt: ${e.message}`);
			}
		}
	} finally {
		sendetGerade = false;
		await ladeAusgang();
	}
}

window.addEventListener('schaden:abgemeldet', () => {
	zustand.benutzer = null;
});
