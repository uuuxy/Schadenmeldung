/**
 * Eine Snackbar für die ganze App (M3: kurze Rückmeldung unten, höchstens eine Aktion, immer nur
 * eine zugleich). Ohne Aktion geht sie nach 5 Sekunden von selbst. Mit Aktion („Rückgängig“) bleibt
 * sie, bis jemand tippt, sie schließt oder eine neue sie ersetzt: Wer langsamer liest oder tippt,
 * soll das Rückgängig nicht verpassen. Mit bleibt steht auch ein Hinweis ohne Aktion so lange („App
 * aktualisiert“), denn den soll niemand verpassen, der gerade nicht hinsieht.
 */
export const snackbar = $state({ text: '', aktion: null, bleibt: false, nummer: 0 });

let uhr;

export function zeige(text, aktion = null, bleibt = aktion !== null) {
	clearTimeout(uhr);
	snackbar.text = text;
	snackbar.aktion = aktion;
	snackbar.bleibt = bleibt;
	snackbar.nummer += 1;
	if (!bleibt) uhr = setTimeout(schliesse, 5000);
}

export function schliesse() {
	clearTimeout(uhr);
	snackbar.text = '';
	snackbar.aktion = null;
	snackbar.bleibt = false;
}
