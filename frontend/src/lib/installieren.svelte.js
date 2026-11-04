/**
 * Als App installieren. Chrome und Edge (Android, Windows) melden mit beforeinstallprompt, dass sie
 * es anbieten; das Ereignis kommt gleich beim Laden und wird hier aufgehoben, bis jemand tippt.
 * Safari auf dem iPhone kennt es nicht, dort geht es nur über „Teilen → Zum Home-Bildschirm“.
 */
export const installieren = $state({ angebot: null });

const SCHLUESSEL = 'schaden.installHinweis';

window.addEventListener('beforeinstallprompt', (e) => {
	e.preventDefault();
	installieren.angebot = e;
});
window.addEventListener('appinstalled', () => (installieren.angebot = null));

/** Läuft die Schadenmeldung schon als installierte App? */
export const alsApp = () => matchMedia('(display-mode: standalone)').matches || navigator.standalone === true;

/** iPhone und iPad; neuere iPads geben sich als Mac aus, haben aber einen Touchscreen. */
export const apfelGeraet = () => /iPhone|iPad|iPod/.test(navigator.userAgent) || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);

export function hinweisWeg() {
	try {
		return localStorage.getItem(SCHLUESSEL) === 'weg';
	} catch {
		return false;
	}
}

export function hinweisAusblenden() {
	try {
		localStorage.setItem(SCHLUESSEL, 'weg');
	} catch {
		/* Speicher gesperrt: Dann kommt der Hinweis eben beim nächsten Mal wieder. */
	}
}
