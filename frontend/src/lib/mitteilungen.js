import { api } from './api.js';
import { apfelGeraet, alsApp } from './installieren.svelte.js';

/**
 * Mitteilungen aufs Handy (push.go): je Gerät ein Abo beim Push-Dienst des Browsers. Auf dem iPhone geht
 * das nur in der App vom Home-Bildschirm; der Schalter sagt das, statt einfach zu fehlen. Die Zahl am
 * App-Symbol setzt der Service Worker mit jeder Mitteilung, beim Öffnen der App verschwindet sie.
 */

const moeglich = () => 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;

async function aboHier() {
	const reg = await navigator.serviceWorker.getRegistration();
	return (await reg?.pushManager.getSubscription()) ?? null;
}

/**
 * Wie es auf diesem Gerät steht: 'an', 'aus', 'gesperrt', 'homescreen' (iPhone im Browser), 'veraltet'
 * (iPhone vor iOS 16.4: Apple lässt Web-Apps dort keine Mitteilungen) oder '' (geht nicht).
 */
export async function mitteilungenStand() {
	if (!moeglich()) return apfelGeraet() ? (alsApp() ? 'veraltet' : 'homescreen') : '';
	if (Notification.permission === 'denied') return 'gesperrt';
	return Notification.permission === 'granted' && (await aboHier()) ? 'an' : 'aus';
}

/** Schaltet Mitteilungen auf diesem Gerät ein; das Handy fragt dabei einmal nach der Erlaubnis. */
export async function mitteilungenEin() {
	// Zuerst die Erlaubnis: Das iPhone fragt nur direkt nach einem Tipp.
	if ((await Notification.requestPermission()) !== 'granted') return false;
	const { schluessel } = await api('/api/push');
	const reg = await navigator.serviceWorker.ready;
	// Erreicht der Browser seinen Push-Dienst nicht, hinge das sonst ohne Ende. Das erste Abo eines Browsers
	// braucht aber oft eine halbe Minute (gemessen mit Chrome: 4 bis 29 Sekunden), darum erst nach einer
	// Minute ehrlich melden, dass es nicht geklappt hat.
	const frist = new Promise((_, nein) => setTimeout(() => nein(new Error('Frist')), 60_000));
	const abo = await Promise.race([reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: alsBytes(schluessel) }), frist]);
	await meldeAn(abo);
	return true;
}

/**
 * Meldet das Abo beim Server an. Lehnt er es ab, gilt es auch hier nicht mehr: Sonst stünde der Schalter
 * auf „an“, und trotzdem käme nichts. Ohne Netz, abgemeldet, gebremst oder bei gestörtem Server bleibt es;
 * beim nächsten Öffnen der App meldet es sich erneut (gesehen).
 */
async function meldeAn(abo) {
	try {
		await api('/api/push', { methode: 'POST', json: abo.toJSON() });
	} catch (e) {
		if (e.status >= 400 && e.status < 500 && e.status !== 401 && e.status !== 429) await abo.unsubscribe().catch(() => {});
		throw e;
	}
}

export async function mitteilungenAus() {
	const abo = await aboHier();
	if (!abo) return;
	await api('/api/push', { methode: 'DELETE', json: { endpoint: abo.endpoint } }).catch(() => {});
	await abo.unsubscribe();
}

/**
 * Beim Öffnen der App: Zahl am Symbol und offene Mitteilungen verschwinden, und das Abo meldet sich beim
 * Server. Das hält es frisch und setzt dort die Zählung der neuen Meldungen zurück.
 */
export async function gesehen() {
	if (!moeglich()) return;
	navigator.clearAppBadge?.()?.catch(() => {});
	const reg = await navigator.serviceWorker.getRegistration();
	for (const n of (await reg?.getNotifications?.()) ?? []) n.close();
	const abo = Notification.permission === 'granted' ? await aboHier() : null;
	if (abo) await meldeAn(abo).catch(() => {});
}

function alsBytes(text) {
	return Uint8Array.from(atob(text.replace(/-/g, '+').replace(/_/g, '/')), (z) => z.charCodeAt(0));
}
