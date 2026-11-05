// Der Ausgang: Meldungen (samt Foto) und Änderungen am Stand (art: 'stand'), die im Funkloch
// abgeschickt wurden. Sie liegen im Browser (IndexedDB) und gehen der Reihe nach raus, sobald wieder
// Verbindung besteht (nachreichen.js).
const DATENBANK = 'schadenmeldung';
const ABLAGE = 'ausgang';

function oeffne() {
	return new Promise((fertig, fehler) => {
		const anfrage = indexedDB.open(DATENBANK, 1);
		anfrage.onupgradeneeded = () => anfrage.result.createObjectStore(ABLAGE, { keyPath: 'id', autoIncrement: true });
		anfrage.onsuccess = () => fertig(anfrage.result);
		anfrage.onerror = () => fehler(anfrage.error);
	});
}

async function mitAblage(modus, arbeit) {
	const db = await oeffne();
	try {
		return await new Promise((fertig, fehler) => {
			const tx = db.transaction(ABLAGE, modus);
			const anfrage = arbeit(tx.objectStore(ABLAGE));
			tx.oncomplete = () => fertig(anfrage.result);
			tx.onerror = () => fehler(tx.error);
			tx.onabort = () => fehler(tx.error);
		});
	} finally {
		db.close();
	}
}

export const legeAb = (eintrag) => mitAblage('readwrite', (a) => a.add(eintrag));
export const alleImAusgang = () => mitAblage('readonly', (a) => a.getAll());
export const entferne = (id) => mitAblage('readwrite', (a) => a.delete(id));
export const leereAusgang = () => mitAblage('readwrite', (a) => a.clear());

/** Nimmt einen Eintrag heraus, solange er noch wartet; undefined, wenn er schon hinaus ist. */
export const nimmHeraus = (id) =>
	mitAblage('readwrite', (a) => {
		const anfrage = a.get(id);
		anfrage.onsuccess = () => anfrage.result && a.delete(id);
		return anfrage;
	});

/** Zufällige Kennung je Meldung: Kommt sie doppelt an (Antwort im Funkloch verloren), legt der Server sie nur einmal an. */
export const neueKennung = () => [...crypto.getRandomValues(new Uint8Array(16))].map((b) => b.toString(16).padStart(2, '0')).join('');
