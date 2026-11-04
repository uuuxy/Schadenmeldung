/**
 * Als App installieren. Chrome und Edge (Android, Windows) melden mit beforeinstallprompt, dass sie
 * es anbieten; das Ereignis kommt gleich beim Laden und wird hier aufgehoben, bis jemand tippt.
 * Auf dem iPhone geht es in jedem Browser nur von Hand über „Teilen → Zum Home-Bildschirm“, denn Apple
 * lässt keine Seite selbst installieren; Firefox auf Android bietet es nur in seinem Menü an.
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

/** Firefox auf Android kann installieren, sagt es der Seite aber nicht (kein beforeinstallprompt). */
export const firefoxAndroid = () => /Android/.test(navigator.userAgent) && /Firefox\//.test(navigator.userAgent);

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
