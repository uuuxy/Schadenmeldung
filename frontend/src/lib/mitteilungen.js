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

/** Wie es auf diesem Gerät steht: 'an', 'aus', 'gesperrt', 'homescreen' (iPhone im Browser) oder '' (geht nicht). */
export async function mitteilungenStand() {
	if (!moeglich()) return apfelGeraet() && !alsApp() ? 'homescreen' : '';
	if (Notification.permission === 'denied') return 'gesperrt';
	return Notification.permission === 'granted' && (await aboHier()) ? 'an' : 'aus';
}

/** Schaltet Mitteilungen auf diesem Gerät ein; das Handy fragt dabei einmal nach der Erlaubnis. */
export async function mitteilungenEin() {
	// Zuerst die Erlaubnis: Das iPhone fragt nur direkt nach einem Tipp.
	if ((await Notification.requestPermission()) !== 'granted') return false;
	const { schluessel } = await api('/api/push');
	const reg = await navigator.serviceWorker.ready;
	// Erreicht der Browser seinen Push-Dienst nicht, hinge das sonst ohne Ende; nach 15 Sekunden lieber
	// ehrlich melden, dass es nicht geklappt hat.
	const frist = new Promise((_, nein) => setTimeout(() => nein(new Error('Frist')), 15_000));
	const abo = await Promise.race([reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: alsBytes(schluessel) }), frist]);
	await api('/api/push', { methode: 'POST', json: abo.toJSON() });
	return true;
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
	if (abo) await api('/api/push', { methode: 'POST', json: abo.toJSON() }).catch(() => {});
}

function alsBytes(text) {
	return Uint8Array.from(atob(text.replace(/-/g, '+').replace(/_/g, '/')), (z) => z.charCodeAt(0));
}
