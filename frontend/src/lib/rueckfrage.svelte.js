/**
 * Rückfrage vor Endgültigem (M3 Dialog: „purposefully interruptive“, nur für riskante Aktionen wie
 * Löschen, höchstens zwei Aktionen). Ein Dialog für die ganze App, angezeigt von Rueckfrage.svelte:
 * `if (await frage({ titel, text, ja: 'Löschen' })) …`
 */
export const rueckfrage = $state({ offen: false, titel: '', text: '', ja: '' });

let antwort = null;

export function frage({ titel, text = '', ja }) {
	antwort?.(false);
	Object.assign(rueckfrage, { offen: true, titel, text, ja });
	return new Promise((fertig) => (antwort = fertig));
}

export function antworte(wert) {
	rueckfrage.offen = false;
	antwort?.(wert);
	antwort = null;
}
