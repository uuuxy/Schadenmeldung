// Der Ausgang: Meldungen, die im Funkloch abgeschickt wurden. Sie liegen samt Foto im Browser
// (IndexedDB) und gehen raus, sobald wieder Verbindung besteht (zustand.svelte.js, sendeAusgang).
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

/** Zufällige Kennung je Meldung: Kommt sie doppelt an (Antwort im Funkloch verloren), legt der Server sie nur einmal an. */
export const neueKennung = () => [...crypto.getRandomValues(new Uint8Array(16))].map((b) => b.toString(16).padStart(2, '0')).join('');
