/**
 * Eine Snackbar für die ganze App (M3: kurze Rückmeldung unten, höchstens eine Aktion, immer nur
 * eine zugleich). Ohne Aktion geht sie nach 5 Sekunden von selbst. Mit Aktion („Rückgängig“) bleibt
 * sie, bis jemand tippt, sie schließt oder eine neue sie ersetzt: Wer langsamer liest oder tippt,
 * soll das Rückgängig nicht verpassen.
 */
export const snackbar = $state({ text: '', aktion: null, nummer: 0 });

let uhr;

export function zeige(text, aktion = null) {
	clearTimeout(uhr);
	snackbar.text = text;
	snackbar.aktion = aktion;
	snackbar.nummer += 1;
	if (!aktion) uhr = setTimeout(schliesse, 5000);
}

export function schliesse() {
	clearTimeout(uhr);
	snackbar.text = '';
	snackbar.aktion = null;
}
