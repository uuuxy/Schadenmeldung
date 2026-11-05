import { chromium } from 'playwright-core';
// Funkloch beim Hausmeister im Handyformat: Ein Stand, der ohne Netz gesetzt wird, bleibt stehen, auch
// nach dem Neuladen, steht als „wartet auf Netz“ in der Liste und geht raus, sobald wieder Netz da ist.
// „Rückgängig“ nimmt eine wartende Änderung zurück, ohne dass etwas beim Server ankommt; Abmelden fragt
// nach, solange etwas wartet. Was wartet, geht nur unter der Person raus, die es gespeichert hat. Eine
// Meldung ohne Netz zeigt „Gespeichert“ und nach dem Nachreichen „Gemeldet – danke!“.
// Braucht den lokalen Server mit Testdaten aus e2e/testdaten.sh.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const H = { 'X-Schaden': '1' };
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };
async function angemeldet(email, optionen = {}) {
	const k = await browser.newContext(optionen);
	await k.request.post(`${B}/api/anmelden`, { headers: H, data: { email, passwort: 'x' } });
	return k;
}

// Zwei eigene Meldungen, damit der Test nicht an den übrigen Testdaten hängt; geprüft wird am Server
// über eine Admin-Sitzung, die vom Funkloch des Handys nichts merkt (Schritt 6 beendet die des Hausmeisters).
const lehrer = await angemeldet('m.mustermann');
const raum = (await (await lehrer.request.get(`${B}/api/raeume`)).json())[0];
const melde = async (was) => (await lehrer.request.post(`${B}/api/meldungen`, { headers: H, multipart: { raumId: String(raum.id), kategorie: 'licht', was } })).json();
const keller = await melde('Funkloch Keller');
const flur = await melde('Funkloch Flur');
const pruefer = await angemeldet('admin@schule.de');
const standAmServer = async (id) => {
	const d = await (await pruefer.request.get(`${B}/api/meldungen`)).json();
	return [...d.offen, ...d.erledigt].find((m) => m.id === id)?.status;
};

const k = await angemeldet('hausmeister', { viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' });
const s = await k.newPage();
s.on('pageerror', (e) => fehler.push(String(e)));
await s.goto(B + '/');
await s.getByRole('group', { name: 'Ansicht' }).waitFor();
// Ohne Service Worker startete die App im Funkloch gar nicht erst.
await s.evaluate(() => navigator.serviceWorker.ready);
const zeile = (was) => s.getByRole('region', { name: 'Meldungen' }).getByRole('listitem').filter({ hasText: was });
await zeile('Funkloch Keller').waitFor();

// 1. Ohne Netz abhaken: Der Stand bleibt stehen und wartet im Ausgang.
await k.setOffline(true);
await s.getByRole('button', { name: `Erledigt: ${raum.name}, Funkloch Keller` }).click();
await s.getByText(`Erledigt: ${raum.name} · wird nachgereicht`).waitFor({ timeout: 5000 });
// Die Zeile gleitet noch hinaus, während die Snackbar schon steht.
const weg = await zeile('Funkloch Keller').waitFor({ state: 'detached', timeout: 3000 }).then(() => true, () => false);
pruefe(weg, 'Ohne Netz abgehakt: nicht mehr unter „Offen“, die Snackbar sagt „wird nachgereicht“');

// 2. Neu laden ohne Netz: Der Stand bleibt, die Zeile sagt „wartet auf Netz“, beim Server ist nichts.
await s.reload();
await s.getByRole('tab', { name: /^Erledigt/ }).click();
await zeile('Funkloch Keller').waitFor({ timeout: 5000 });
pruefe((await zeile('Funkloch Keller').getByText('wartet auf Netz').count()) === 1, 'Nach dem Neuladen ohne Netz: unter „Erledigt“ mit „wartet auf Netz“');
if (process.argv[2]) await s.screenshot({ path: `${process.argv[2]}/funkloch-wartet.png` });
pruefe((await standAmServer(keller.id)) === 'offen', 'Der Server hat noch nichts bekommen');

// 3. „Rückgängig“ ohne Netz: Die wartende Änderung fällt weg.
await s.getByRole('tab', { name: /^Offen/ }).click();
await zeile('Funkloch Flur').getByRole('button').first().click();
await s.getByRole('group', { name: 'Stand' }).getByRole('button', { name: 'In Arbeit' }).click();
await s.getByText(`In Arbeit: ${raum.name} · wird nachgereicht`).waitFor({ timeout: 5000 });
pruefe((await zeile('Funkloch Flur').getByText('wartet auf Netz').count()) === 1, '„In Arbeit“ ohne Netz: zurück in der Liste, „wartet auf Netz“');
await s.getByRole('button', { name: 'Rückgängig' }).click();
await s.waitForTimeout(300);
pruefe((await zeile('Funkloch Flur').getByText(/wartet auf Netz|In Arbeit/).count()) === 0, '„Rückgängig“ ohne Netz nimmt die wartende Änderung zurück');

// 4. Abmelden ohne Netz, solange „Erledigt“ wartet: Die App fragt nach, statt still zu löschen.
await s.getByRole('button', { name: 'Konto und Einstellungen' }).click();
await s.getByRole('button', { name: 'Abmelden', exact: true }).click();
const dialog = s.getByRole('dialog');
const gefragt = await dialog.getByText('1 Änderung ist noch nicht beim Server angekommen', { exact: false }).waitFor({ timeout: 5000 }).then(() => true, () => false);
pruefe(gefragt, 'Abmelden ohne Netz fragt nach, solange eine Änderung wartet');
if (process.argv[2]) await s.screenshot({ path: `${process.argv[2]}/funkloch-abmelden.png` });
await dialog.getByRole('button', { name: 'Abbrechen' }).click();
pruefe(await s.getByRole('button', { name: 'Konto und Einstellungen' }).isVisible(), '„Abbrechen“: weiter angemeldet, die Änderung wartet weiter');

// 5. Wieder Netz: Was wartet, geht von selbst raus.
await k.setOffline(false);
await s.getByText('Änderung nachgereicht').waitFor({ timeout: 10000 });
pruefe((await standAmServer(keller.id)) === 'erledigt', 'Wieder Netz: „Erledigt“ ist beim Server angekommen');
pruefe((await standAmServer(flur.id)) === 'offen', 'Die zurückgenommene Änderung kam nie beim Server an');

// 6. Endet die Sitzung, während etwas wartet (etwa „Auf allen Geräten abmelden“), und meldet sich danach
// eine andere Person am Gerät an, geht es nicht unter deren Namen raus, sondern erst mit dem Hausmeister.
const anmelden = async (name) => {
	await s.getByLabel(/Name oder E-Mail|E-Mail der Schule/).fill(name);
	await s.getByLabel('Passwort des Schul-Postfachs').fill('x');
	await s.getByRole('button', { name: 'Anmelden' }).click();
	await s.getByRole('button', { name: 'Konto und Einstellungen' }).waitFor();
};
const netzZurueck = () => s.evaluate(() => window.dispatchEvent(new Event('online')));
await k.setOffline(true);
await s.getByRole('button', { name: `Erledigt: ${raum.name}, Funkloch Flur` }).click();
await s.getByText(`Erledigt: ${raum.name} · wird nachgereicht`).waitFor({ timeout: 5000 });
await (await angemeldet('hausmeister')).request.post(`${B}/api/abmelden?ueberall=1`, { headers: H });
await k.setOffline(false);
await anmelden('direktorin');
await netzZurueck();
await s.waitForTimeout(1500);
pruefe((await standAmServer(flur.id)) === 'offen', 'Nach beendeter Sitzung meldet sich die Leitung an: Die Änderung des Hausmeisters geht nicht unter ihrem Namen raus');
await s.getByRole('button', { name: 'Konto und Einstellungen' }).click();
await s.getByRole('button', { name: 'Abmelden', exact: true }).click();
await anmelden('hausmeister');
await netzZurueck();
await s.getByText('Änderung nachgereicht').waitFor({ timeout: 10000 });
pruefe((await standAmServer(flur.id)) === 'erledigt', 'Der Hausmeister wieder angemeldet: Jetzt geht seine Änderung raus');

// 7. Eine Meldung ohne Netz: Die Bestätigung sagt „Gespeichert“ und springt auf „Gemeldet – danke!“,
// sobald die Meldung nachgereicht ist.
const qr = (await (await pruefer.request.get(`${B}/api/raeume`)).json()).find((r) => r.id === raum.id).code;
await s.goto(`${B}/r/${qr}`);
await s.getByRole('heading', { name: 'Schaden melden' }).waitFor();
await s.getByRole('button', { name: 'Wasser & WC' }).click();
await s.getByRole('button', { name: 'Wasserhahn tropft' }).click();
await k.setOffline(true);
await s.getByRole('button', { name: /^Melden · / }).click();
await s.getByRole('heading', { name: 'Gespeichert' }).waitFor({ timeout: 5000 });
await k.setOffline(false);
const umgesprungen = await s.getByRole('heading', { name: 'Gemeldet – danke!' }).waitFor({ timeout: 10000 }).then(() => true, () => false);
pruefe(umgesprungen, 'Meldung ohne Netz: Nach dem Nachreichen springt „Gespeichert“ auf „Gemeldet – danke!“');
const nachgereicht = (await (await pruefer.request.get(`${B}/api/meldungen`)).json()).offen
	.filter((m) => m.raumId === raum.id && m.was === 'Wasserhahn tropft')
	.sort((x, y) => y.id - x.id)[0];

for (const m of [keller, flur, nachgereicht].filter(Boolean)) await pruefer.request.delete(`${B}/api/meldungen/${m.id}`, { headers: H });
await browser.close();
if (fehler.length) {
	console.log(fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
