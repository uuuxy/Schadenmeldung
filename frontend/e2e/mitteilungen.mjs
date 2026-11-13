import { chromium, request } from 'playwright-core';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
// Mitteilungen echt bis zum Push-Dienst: Das installierte Google Chrome holt sich ein Abo bei Googles
// Dienst. Das Playwright-Chromium und Chrome ohne Fenster können das nicht, darum geht ein Fenster auf.
// Geprüft wird: Lehnt der Server ein Gerät ab, steht der Schalter auf „aus“; jede neue Meldung kommt als
// Mitteilung an, und die Zahl am App-Symbol zählt hoch; beim Öffnen der App verschwinden beide, und die
// Zählung beginnt neu; ausgeschaltet kommt nichts mehr. Braucht Google Chrome, Netz und den lokalen Server
// mit Testdaten aus e2e/testdaten.sh.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const H = { 'X-Schaden': '1' };
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };

const profil = mkdtempSync(join(tmpdir(), 'schaden-push-'));
const k = await chromium.launchPersistentContext(profil, { channel: 'chrome', headless: false, permissions: ['notifications'], viewport: { width: 390, height: 844 } });
await k.request.post(`${B}/api/anmelden`, { headers: H, data: { email: 'hausmeister', passwort: 'x' } });
const s = k.pages()[0] ?? (await k.newPage());
s.on('pageerror', (e) => fehler.push(String(e)));
await s.goto(B + '/');
await s.evaluate(() => navigator.serviceWorker.ready);

// Gemeldet wird von einer Lehrerin über die Schnittstelle, aufgeräumt über eine Admin-Sitzung.
const sitzung = async (email) => {
	const r = await request.newContext({ baseURL: B, extraHTTPHeaders: H });
	await r.post('/api/anmelden', { data: { email, passwort: 'x' } });
	return r;
};
const lehrerin = await sitzung('s.schmidt');
const admin = await sitzung('admin@schule.de');
const raum = (await (await lehrerin.get('/api/raeume')).json())[0];
const gemeldet = [];
async function melde(was, dringend = false) {
	const m = await (await lehrerin.post('/api/meldungen', { multipart: { raumId: String(raum.id), kategorie: 'licht', was, ...(dringend && { dringend: '1' }) } })).json();
	gemeldet.push(m.id);
	return m;
}

// Was der Browser zeigt und hat, gefragt im Fenster der App.
const mitteilungen = () => s.evaluate(async () => (await (await navigator.serviceWorker.ready).getNotifications()).map((n) => ({ titel: n.title, text: n.body, weg: n.data?.weg })));
const aboDa = () => s.evaluate(async () => !!(await (await navigator.serviceWorker.ready).pushManager.getSubscription()));
async function warte(bedingung, frist = 30000) {
	for (const ende = Date.now() + frist; Date.now() < ende; await s.waitForTimeout(400)) if (await bedingung()) return true;
	return false;
}
// Gibt zurück, was angekommen ist, sobald es da ist: Ein Klick auf die Mitteilung am Bildschirm schließt sie.
async function warteAufMitteilungen(anzahl, frist) {
	let liste = [];
	await warte(async () => (liste = await mitteilungen()).length >= anzahl, frist);
	return liste;
}
// Wie am Handy: Die App geht in den Hintergrund und kommt zurück.
async function wiederGeoeffnet() {
	for (const z of ['hidden', 'visible']) {
		await s.evaluate((zustand) => {
			Object.defineProperty(document, 'visibilityState', { value: zustand, configurable: true });
			document.dispatchEvent(new Event('visibilitychange'));
		}, z);
	}
	await s.waitForTimeout(1500);
}

const konto = s.getByRole('button', { name: 'Konto und Einstellungen' });
const schalter = s.getByRole('switch', { name: /^Auf diesem Gerät/ });
const schalterAn = async () => (await schalter.getAttribute('aria-checked')) === 'true';
async function imMenue(tun) {
	await konto.click();
	await schalter.waitFor();
	await s.waitForTimeout(400);
	const ergebnis = await tun();
	await s.keyboard.press('Escape');
	return ergebnis;
}
// Lässt den Server ein Abo bei einem fremden Push-Dienst sehen: Der echte Server lehnt es dann ab.
const fremderDienst = async (route) => {
	if (route.request().method() !== 'POST') return route.continue();
	await route.continue({ postData: JSON.stringify({ ...route.request().postDataJSON(), endpoint: 'https://boese.example/push/x' }) });
};

// 1. Einschalten mit echtem Abo bei Googles Push-Dienst, wie im Alltag zuerst. Das erste Abo eines
//    Browsers dauert oft eine halbe Minute; so lange sagt der Schalter, dass es läuft.
pruefe(!(await imMenue(schalterAn)), 'Zu Beginn: Schalter „Auf diesem Gerät“ aus');
await imMenue(async () => {
	await schalter.click();
	pruefe(await s.getByText('wird eingeschaltet …').waitFor({ timeout: 2000 }).then(() => true, () => false), 'Beim Einschalten steht „wird eingeschaltet …“ da');
	pruefe((await warte(aboDa, 70000)) && (await warte(schalterAn, 5000)), 'Einschalten: echtes Abo, Schalter bleibt auf „an“');
});

// 2. Neue Meldungen kommen über Google an; die Zahl am Symbol zählt mit (im Service Worker mitgeschrieben).
const zaehleMit = () => {
	self.zahlen = [];
	self.zeiten = [];
	const echt = self.navigator.setAppBadge?.bind(self.navigator);
	self.navigator.setAppBadge = (n) => (self.zahlen.push(n), self.zeiten.push(Date.now()), echt ? echt(n) : Promise.resolve());
};
const zahlenVon = (worker) => worker.evaluate(() => self.zahlen ?? null).catch(() => null);
const sw = k.serviceWorkers()[0] ?? (await k.waitForEvent('serviceworker'));
await sw.evaluate(zaehleMit);
const zahlen = () => zahlenVon(sw);
const gemeldetUm = Date.now();
const erste = await melde('Push-Probe Lampe');
// Die erste Mitteilung an ein neues Abo braucht bei Google manchmal länger.
const [n1] = await warteAufMitteilungen(1, 180000);
const dauer = await sw.evaluate((t) => (self.zeiten?.[0] ? ((self.zeiten[0] - t) / 1000).toFixed(1) + ' s' : '–'), gemeldetUm).catch(() => '–');
pruefe(n1?.titel === 'Neue Meldung' && n1.text === `Push-Probe Lampe – ${erste.ort}` && n1.weg === `/meldung/${erste.id}`, `Mitteilung kommt an (nach ${dauer}): „${n1?.titel}: ${n1?.text}“, führt zu /meldung/${erste.id}`);
await melde('Push-Probe Wasser', true);
const titel = (await warteAufMitteilungen(2, 120000)).map((n) => n.titel);
pruefe(titel.includes('Dringende Meldung'), 'Dringendes heißt „Dringende Meldung“');
const z = await zahlen();
pruefe(JSON.stringify(z) === '[1,2]', `Zahl am App-Symbol zählt hoch: ${JSON.stringify(z)}`);

// 3. Beim Öffnen der App verschwinden Zahl und Mitteilungen, und die Zählung beginnt neu.
await s.evaluate(() => {
	window.geleert = 0;
	const echt = navigator.clearAppBadge?.bind(navigator);
	navigator.clearAppBadge = () => (window.geleert++, echt ? echt() : Promise.resolve());
});
await wiederGeoeffnet();
pruefe((await s.evaluate(() => window.geleert)) >= 1 && (await mitteilungen()).length === 0, 'App geöffnet: Zahl gelöscht, Mitteilungen weg');
await melde('Push-Probe Fenster');
await warteAufMitteilungen(1, 120000);
const z2 = await zahlen();
pruefe(z2?.at(-1) === 1, `Danach zählt sie wieder ab 1: ${JSON.stringify(z2)}`);

// 3b. Jedes Gerät zählt für sich: Eine zweite Hausmeisterin mit eigenem Gerät bekommt dieselben Meldungen;
//     wer die App öffnet, setzt nur die Zahl auf dem eigenen Gerät zurück. Eingeschaltet wie in der App.
await admin.post('/api/rollen', { data: { email: 'hausmeisterin', rolle: 'hausmeister' } });
const profil2 = mkdtempSync(join(tmpdir(), 'schaden-push-'));
const k2 = await chromium.launchPersistentContext(profil2, { channel: 'chrome', headless: false, permissions: ['notifications'], viewport: { width: 390, height: 844 } });
await k2.request.post(`${B}/api/anmelden`, { headers: H, data: { email: 'hausmeisterin', passwort: 'x' } });
const s2 = k2.pages()[0] ?? (await k2.newPage());
await s2.goto(B + '/');
const angenommen = await s2.evaluate(async () => {
	const { schluessel } = await (await fetch('/api/push')).json();
	const key = Uint8Array.from(atob(schluessel.replace(/-/g, '+').replace(/_/g, '/')), (z) => z.charCodeAt(0));
	const abo = await (await navigator.serviceWorker.ready).pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: key });
	return (await fetch('/api/push', { method: 'POST', headers: { 'Content-Type': 'application/json', 'X-Schaden': '1' }, body: JSON.stringify(abo.toJSON()) })).status;
});
const sw2 = k2.serviceWorkers()[0] ?? (await k2.waitForEvent('serviceworker'));
await sw2.evaluate(zaehleMit);
const naechste = async (worker, anzahl) => (await warte(async () => ((await zahlenVon(worker))?.length ?? 0) >= anzahl, 180000), (await zahlenVon(worker))?.at(-1));
const bisherA = (await zahlen())?.length ?? 0;
await melde('Push-Probe Tür');
const [a1, b1] = [await naechste(sw, bisherA + 1), await naechste(sw2, 1)];
await wiederGeoeffnet();
await melde('Push-Probe Heizung');
const [a2, b2] = [await naechste(sw, bisherA + 2), await naechste(sw2, 2)];
pruefe(angenommen === 204 && a1 === 2 && b1 === 1 && a2 === 1 && b2 === 2, `Jedes Gerät zählt für sich: Hausmeister ${a1} → App geöffnet → ${a2}, Hausmeisterin ${b1} → ${b2}`);
await k2.close();
rmSync(profil2, { recursive: true, force: true });
await admin.delete('/api/rollen/hausmeisterin@schule.de');

// 4. Lehnt der Server das Gerät beim Öffnen ab, gilt das Abo auch im Browser nicht mehr.
await s.route(`${B}/api/push`, fremderDienst);
await wiederGeoeffnet();
pruefe(!(await aboDa()) && !(await imMenue(schalterAn)), 'Abgelehnt beim Öffnen: kein Abo mehr, Schalter „aus“');

// 5. Lehnt der Server das Gerät beim Einschalten ab: Die Snackbar sagt es, der Schalter springt auf „aus“.
await imMenue(async () => {
	await schalter.click();
	const gesagt = await s.getByText('Dieses Gerät kann keine Mitteilungen empfangen.').waitFor({ timeout: 70000 }).then(() => true, () => false);
	await s.waitForTimeout(500);
	pruefe(gesagt && !(await schalterAn()) && !(await aboDa()), 'Server lehnt ab: Snackbar sagt es, Schalter springt auf „aus“, kein Abo im Browser');
});
await s.unroute(`${B}/api/push`, fremderDienst);

// 6. Wieder an und dann aus: Danach bringt eine neue Meldung keine Mitteilung mehr.
await imMenue(async () => {
	await schalter.click();
	await warte(aboDa);
	await warte(schalterAn, 5000);
	await schalter.click();
	await warte(async () => !(await aboDa()));
});
await wiederGeoeffnet();
await melde('Push-Probe Stille');
await s.waitForTimeout(8000);
pruefe(!(await aboDa()) && (await mitteilungen()).length === 0, 'Ausgeschaltet: keine Mitteilung mehr');

for (const id of gemeldet) await admin.delete(`/api/meldungen/${id}`);
await k.close();
rmSync(profil, { recursive: true, force: true });
if (fehler.length) {
	console.log(fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
