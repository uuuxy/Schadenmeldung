// Service Worker der Schadensmeldung: Die App startet sofort und auch ohne Netz.
// Seiten kommen zuerst vom Netz (ein Update wirkt sofort) und nur im Funkloch aus dem Speicher. Die
// gehashten Dateien unter /assets/ ändern sich nie und kommen gleich aus dem Speicher. API und Fotos
// laufen nie über den Speicher: Sie sind persönlich und müssen aktuell sein.
const SEITE = 'schaden-seite-v2';
const DATEIEN = 'schaden-dateien-v2';
const GRUND = ['/', '/darstellung.js', '/manifest.webmanifest', '/icon.svg'];

self.addEventListener('install', (e) => {
	e.waitUntil(caches.open(SEITE).then((c) => c.addAll(GRUND)).then(() => self.skipWaiting()));
});

self.addEventListener('activate', (e) => {
	const bekannt = [SEITE, DATEIEN];
	e.waitUntil(
		caches
			.keys()
			.then((namen) => Promise.all(namen.filter((n) => !bekannt.includes(n)).map((n) => caches.delete(n))))
			.then(() => self.clients.claim())
	);
});

function mitFrist(versprechen, ms) {
	return Promise.race([versprechen, new Promise((_, nein) => setTimeout(() => nein(new Error('Frist')), ms))]);
}

async function seite(e) {
	const speicher = await caches.open(SEITE);
	try {
		// Bei schwachem WLAN nicht ewig warten: nach 4 Sekunden lieber den gespeicherten Stand zeigen.
		const antwort = await mitFrist(fetch(e.request), 4000);
		if (antwort.ok) e.waitUntil(speicher.put('/', antwort.clone()));
		return antwort;
	} catch {
		return (await speicher.match('/')) ?? Response.error();
	}
}

async function datei(e) {
	const speicher = await caches.open(DATEIEN);
	const gespeichert = await speicher.match(e.request);
	if (gespeichert) return gespeichert;
	const antwort = await fetch(e.request);
	if (antwort.ok) e.waitUntil(speicher.put(e.request, antwort.clone()));
	return antwort;
}

async function sonstiges(e) {
	const speicher = await caches.open(SEITE);
	try {
		const antwort = await fetch(e.request);
		if (antwort.ok) e.waitUntil(speicher.put(e.request, antwort.clone()));
		return antwort;
	} catch {
		return (await speicher.match(e.request)) ?? Response.error();
	}
}

// Mitteilungen aufs Handy (push.go): Titel, Text, Weg zur Meldung und die Zahl der neuen Meldungen für
// das App-Symbol. Ein Tipp öffnet die Meldung, in einem offenen Fenster der App oder in einem neuen.
self.addEventListener('push', (e) => {
	let n = {};
	try {
		n = e.data?.json() ?? {};
	} catch {
		/* unlesbar: dann eben eine schlichte Mitteilung */
	}
	const zahl = Number(n.zahl) || 0;
	e.waitUntil(
		Promise.all([
			self.registration.showNotification(n.titel || 'Schadensmeldung', {
				body: n.text || 'Eine neue Meldung ist da.',
				tag: n.tag,
				icon: '/icon-192.png',
				badge: '/badge.png',
				data: { weg: n.weg || '/' }
			}),
			zahl > 0 && self.navigator.setAppBadge ? self.navigator.setAppBadge(zahl).catch(() => {}) : null
		])
	);
});

self.addEventListener('notificationclick', (e) => {
	e.notification.close();
	const ziel = new URL(e.notification.data?.weg || '/', self.location.origin).href;
	e.waitUntil(
		(async () => {
			for (const fenster of await self.clients.matchAll({ type: 'window', includeUncontrolled: true })) {
				try {
					await fenster.focus();
					return await fenster.navigate(ziel);
				} catch {
					/* Dieses Fenster lässt sich nicht lenken: das nächste oder ein neues. */
				}
			}
			return self.clients.openWindow(ziel);
		})()
	);
});

self.addEventListener('fetch', (e) => {
	const url = new URL(e.request.url);
	if (e.request.method !== 'GET' || url.origin !== self.location.origin) return;
	if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/fotos/') || url.pathname === '/healthz') return;
	// Die Frage nach einer neuen Fassung (lib/fassung.js) geht am Speicher vorbei: Sonst läge dort eine
	// index.html, deren Dateien noch fehlen, und die App startete im Funkloch nicht.
	if (e.request.cache === 'no-store') return;
	if (e.request.mode === 'navigate') e.respondWith(seite(e));
	else if (url.pathname.startsWith('/assets/')) e.respondWith(datei(e));
	else e.respondWith(sonstiges(e));
});
