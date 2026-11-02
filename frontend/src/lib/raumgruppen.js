// Für „Wo?“ ordnen sich die Räume in Gruppen, damit nicht alle auf einmal dastehen (Google Kalender
// sucht Räume genauso: Gebäude, dann Raum). Ein gepflegter Bereich (Einstellungen → Räume) gilt
// zuerst. Fehlt er, sagt der Name, wohin der Raum gehört, so wie Schulen ihre Räume meist benennen:
// „C204“ liegt in C, im 2. Stock. Was nicht so heißt („Aula“, „Garten“), steht unter „Weitere Räume“.

export const WEITERE = 'Weitere Räume';

// Buchstaben vorn, dann die Nummer: „C204“, „CU01“, „H 1“, „R12 Chemie“.
const MUSTER = /^([A-Za-zÄÖÜäöü]{1,3}) ?(\d{1,4})\b/;

const natuerlich = (a, b) => a.name.localeCompare(b.name, 'de', { numeric: true, sensitivity: 'base' });
const stockName = (s) => (s === 0 ? 'EG' : `${s}. OG`);

function zerlege(raum) {
	const treffer = MUSTER.exec(raum.name.trim());
	return treffer ? { haus: treffer[1].toUpperCase(), nummer: treffer[2] } : null;
}

/**
 * Teilt die Räume in Gruppen: gepflegte Bereiche in der Reihenfolge der Raumliste, dann die aus den
 * Namen nach dem Alphabet, „Weitere Räume“ zuletzt. Eine Gruppe aus dem Namen mit nur einem Raum
 * fällt zu den weiteren. Tragen alle Räume einer Namensgruppe dreistellige Nummern, gliedert die
 * erste Ziffer sie nach Stockwerken (0 = EG), sonst bleibt `stockwerke` leer.
 * @returns {{ titel: string, raeume: object[], stockwerke: { titel: string, raeume: object[] }[] }[]}
 */
export function raumGruppen(raeume) {
	const gepflegt = new Map();
	const ausNamen = new Map();
	const weitere = [];
	for (const r of raeume) {
		if (r.bereich) {
			if (!gepflegt.has(r.bereich)) gepflegt.set(r.bereich, []);
			gepflegt.get(r.bereich).push(r);
			continue;
		}
		const teile = zerlege(r);
		if (!teile) weitere.push(r);
		else {
			if (!ausNamen.has(teile.haus)) ausNamen.set(teile.haus, []);
			ausNamen.get(teile.haus).push(r);
		}
	}
	const gruppen = [...gepflegt].map(([titel, liste]) => ({ titel, raeume: liste, stockwerke: [] }));
	for (const haus of [...ausNamen.keys()].sort((a, b) => a.localeCompare(b, 'de'))) {
		const liste = ausNamen.get(haus).sort(natuerlich);
		if (liste.length === 1) weitere.push(liste[0]);
		else gruppen.push({ titel: haus, raeume: liste, stockwerke: stockwerkeVon(liste) });
	}
	if (weitere.length) gruppen.push({ titel: WEITERE, raeume: weitere.sort(natuerlich), stockwerke: [] });
	return gruppen;
}

function stockwerkeVon(liste) {
	const nummern = liste.map((r) => zerlege(r).nummer);
	if (!nummern.every((n) => n.length >= 3)) return [];
	const stock = new Map();
	liste.forEach((r, i) => {
		const s = Math.floor(Number(nummern[i]) / 100);
		if (!stock.has(s)) stock.set(s, []);
		stock.get(s).push(r);
	});
	if (stock.size < 2) return [];
	return [...stock].sort((a, b) => a[0] - b[0]).map(([s, raeumeImStock]) => ({ titel: stockName(s), raeume: raeumeImStock }));
}
