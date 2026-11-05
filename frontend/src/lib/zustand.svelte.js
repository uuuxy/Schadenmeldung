import { api } from './api.js';
import { alleImAusgang, leereAusgang } from './ausgang.js';

/** Der gemeinsame Zustand der App: wer angemeldet ist, die Meldungen, die Räume und das Netz. */
export const zustand = $state({
	bereit: false,
	benutzer: null,
	// artenAus: Kacheln, die das Formular nicht zeigt (Einstellungen → Arten), etwa Medientechnik.
	einstellungen: { schulname: '', domain: '', fotoTage: 30, aufbewahrungTage: 365, mailAktiv: false, artenAus: [] },
	offen: [],
	erledigt: [],
	meineErledigt: [],
	mailStoerungSeit: null,
	raeume: [],
	// Funkloch: Die App zeigt den letzten Stand (standVon) und sammelt im Ausgang Meldungen (ausgang)
	// und Änderungen am Stand (wartend), bis wieder Netz da ist.
	offline: false,
	standVon: null,
	ausgang: [],
	wartend: []
});

// Kollegium meldet nur. Hausmeister, Leitung und Admin dürfen alles, auch Räume und Rollen; die
// Verwaltung (Sekretariat) sieht die Übersicht und druckt QR-Aufkleber. Der Server prüft dasselbe.
export const darfBearbeiten = () => ['hausmeister', 'leitung', 'admin'].includes(zustand.benutzer?.rolle);
export const darfSehen = () => ['hausmeister', 'leitung', 'verwaltung', 'admin'].includes(zustand.benutzer?.rolle);

// ── Letzter Stand auf dem Gerät ──────────────────────────────────────────────
// Damit die App im Funkloch nicht leer oder mit der Anmeldung startet. Beim Abmelden gelöscht.
const STAND = 'schaden.stand';

export function merkeStand() {
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
	zustand.wartend = [];
	zustand.offen = [];
	zustand.erledigt = [];
}

// ── Laden ────────────────────────────────────────────────────────────────────

export async function starte() {
	await ladeEinstellungen().catch(() => {
		/* Ohne Einstellungen geht es mit den Vorgaben weiter. */
	});
	// Zuerst der Ausgang: Wartende Änderungen gelten schon, wenn die Meldungen kommen (ladeMeldungen).
	await ladeAusgang();
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
	zustand.bereit = true;
}

export async function ladeEinstellungen() {
	zustand.einstellungen = await api('/api/einstellungen');
}

export async function ladeAlles() {
	await Promise.all([ladeMeldungen(), ladeRaeume()]);
}

export async function ladeMeldungen() {
	const daten = await api('/api/meldungen');
	const { offen, erledigt, meineErledigt = [], mailStoerungSeit = null } = daten;
	Object.assign(zustand, { offen, erledigt, meineErledigt, mailStoerungSeit });
	// Was noch im Ausgang wartet, gilt schon: Sonst spränge der Stand bis zum Nachreichen zurück.
	for (const w of zustand.wartend) {
		const m = zustand.offen.find((x) => x.id === w.meldungId) ?? zustand.erledigt.find((x) => x.id === w.meldungId);
		if (m) uebernimm(mitAenderung(m, w.aenderung, w.erstellt));
	}
	zustand.offline = false;
	zustand.standVon = null;
	merkeStand();
}

export async function ladeRaeume() {
	zustand.raeume = await api('/api/raeume');
	merkeStand();
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

/** Die Meldung, wie sie nach der Änderung aussieht; erledigt bleibt beim ersten Abhaken stehen wie im Server. */
export function mitAenderung(meldung, aenderung, zeit = new Date().toISOString()) {
	return { ...meldung, ...aenderung, ...(aenderung.status === 'erledigt' && meldung.status !== 'erledigt' && { erledigtAm: zeit }) };
}

export async function setzeStatus(id, aenderung) {
	const neu = await api(`/api/meldungen/${id}/status`, { methode: 'POST', json: aenderung });
	uebernimm(neu);
	return neu;
}

// ── Ausgang ──────────────────────────────────────────────────────────────────

/** Liest den Ausgang in den Zustand; was hinein- und hinausgeht, regelt nachreichen.js. */
export async function ladeAusgang() {
	try {
		const alle = await alleImAusgang();
		zustand.ausgang = alle.filter((e) => e.art !== 'stand').map(({ id, anzeige, erstellt }) => ({ id, ...anzeige, erstellt }));
		zustand.wartend = alle.filter((e) => e.art === 'stand');
	} catch {
		zustand.ausgang = [];
		zustand.wartend = [];
	}
}

window.addEventListener('schaden:abgemeldet', () => {
	zustand.benutzer = null;
});
