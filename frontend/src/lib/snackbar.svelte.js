/** Eine Snackbar für die ganze App (M3: kurze Rückmeldung unten, höchstens eine Aktion). */
export const snackbar = $state({ text: '', aktion: null, nummer: 0 });

let uhr;

export function zeige(text, aktion = null) {
	clearTimeout(uhr);
	snackbar.text = text;
	snackbar.aktion = aktion;
	snackbar.nummer += 1;
	// Mit Aktion länger stehen lassen, damit „Rückgängig“ erreichbar bleibt.
	uhr = setTimeout(schliesse, aktion ? 6000 : 3500);
}

export function schliesse() {
	clearTimeout(uhr);
	snackbar.text = '';
	snackbar.aktion = null;
}
