/**
 * Als App installieren. Chrome und Edge (Android, Windows) melden mit beforeinstallprompt, dass sie
 * es anbieten; das Ereignis wird hier aufgehoben, bis jemand tippt. Sonst geht es nur von Hand über
 * das Menü des Browsers, Schritt für Schritt (anleitung), denn Apple lässt keine Seite selbst
 * installieren, und Firefox sagt der Seite nicht, dass er es kann.
 */
export const installieren = $state({ angebot: null });

const SCHLUESSEL = 'schaden.installHinweis';

window.addEventListener('beforeinstallprompt', (e) => {
	e.preventDefault();
	installieren.angebot = e;
});
window.addEventListener('appinstalled', () => (installieren.angebot = null));

/** Läuft die Schadensmeldung schon als installierte App? */
export const alsApp = () => matchMedia('(display-mode: standalone)').matches || navigator.standalone === true;

const ua = navigator.userAgent;
/** Neuere iPads geben sich als Mac aus, haben aber einen Touchscreen. */
const ipad = () => /iPad/.test(ua) || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);

/** iPhone und iPad. */
export const apfelGeraet = () => /iPhone|iPod/.test(ua) || ipad();

// Die Symbole, nach denen man suchen muss; der Name ist für Bildschirmleser.
const teilen = { icon: 'ios_share', name: 'Teilen-Symbol' };
const punkte = { icon: 'more_vert', name: 'drei Punkte' };
const punkteQuer = { icon: 'more_horiz', name: 'drei Punkte' };
const schritt = (text, symbol) => ({ text, ...symbol });

// Ab iOS 26 heißt der Eintrag „Zu Home-Bildschirm hinzufügen“, davor „Zum Home-Bildschirm“.
const homeNeu = schritt('Nach unten wischen und „Zu Home-Bildschirm hinzufügen“ tippen.');
const alsWebApp = schritt('Den Schalter „Als Web-App öffnen“ an lassen und „Hinzufügen“ tippen.');
const hinzufuegen = schritt('„Hinzufügen“ tippen.');

/**
 * Die Schritte für diesen Browser, mit den Namen, die er gerade zeigt; geprüft im November 2026 an den
 * Hilfeseiten von Apple (iOS 26 und 27), Google (Chrome 150) und Mozilla und am Quelltext von Firefox.
 * {icon} im Text steht für das Symbol, nach dem man suchen muss. null: kein bekannter Weg.
 */
export function anleitung() {
	if (apfelGeraet()) {
		const apfel = {
			titel: 'Als App auf den Home-Bildschirm',
			danach: 'Danach die Schadensmeldung über das neue Symbol öffnen und dort anmelden.'
		};
		if (ipad()) {
			return { ...apfel, schritte: [schritt('Oben auf {icon} tippen, bei manchen Ansichten erst auf die drei Punkte.', teilen), schritt('„Zu Home-Bildschirm hinzufügen“ tippen (ältere iPads: „Zum Home-Bildschirm“).'), hinzufuegen] };
		}
		if (/CriOS\//.test(ua)) {
			return { ...apfel, schritte: [schritt('Rechts in der Adresszeile auf {icon} tippen.', teilen), schritt('Nach unten wischen und „Zu Home-Bildschirm hinzufügen“ tippen (ältere iPhones: „Zum Home-Bildschirm“).'), hinzufuegen] };
		}
		const safari = /FxiOS|EdgiOS|OPiOS/.test(ua) ? 0 : Number(ua.match(/Version\/(\d+)/)?.[1] ?? 0);
		// Safari 27 hat „Teilen“ ins Seitenmenü links neben der Adresse gelegt, Safari 26 unter „Mehr“;
		// wer die Leiste auf „Unten“ oder „Oben“ gestellt hat, sieht das Teilen-Symbol gleich.
		const direkt = schritt('Ist das Teilen-Symbol {icon} schon zu sehen: gleich darauf tippen, weiter bei Schritt 3.', { icon: teilen.icon });
		if (safari >= 27) {
			return { ...apfel, hinweis: direkt, schritte: [schritt('Unten auf das Symbol direkt links neben der Adresse tippen (Seitenmenü).'), schritt('„Teilen“ tippen.'), homeNeu, alsWebApp] };
		}
		if (safari === 26) {
			return { ...apfel, hinweis: direkt, schritte: [schritt('Unten rechts auf {icon} tippen.', punkteQuer), schritt('„Teilen“ tippen.'), homeNeu, alsWebApp] };
		}
		if (safari) {
			return { ...apfel, schritte: [schritt('Unten in der Leiste auf {icon} tippen.', teilen), schritt('Nach unten wischen und „Zum Home-Bildschirm“ tippen.'), hinzufuegen] };
		}
		return { ...apfel, schritte: [schritt('Im Menü des Browsers auf „Teilen“ {icon} tippen.', teilen), homeNeu, hinzufuegen] };
	}
	// In Apps eingebettete Browser (Kennung „wv“) können nichts installieren.
	if (!/Android/.test(ua) || /; wv\)/.test(ua)) return null;
	const android = { titel: 'Als App auf den Startbildschirm', danach: 'Danach ist die Schadensmeldung mit einem Tipp auf das neue Symbol da.' };
	if (/Firefox\//.test(ua)) {
		return { ...android, schritte: [schritt('Auf {icon} tippen, oben oder unten rechts.', punkte), schritt('„Mehr“ tippen.'), schritt('„App zum Startbildschirm hinzufügen“ tippen.'), schritt('Im nächsten Fenster „Hinzufügen“ tippen.')] };
	}
	// Nur Chrome selbst: Samsung Internet, Edge und Opera haben eigene Menüs.
	if (/Chrome\//.test(ua) && !/SamsungBrowser|EdgA|OPR|YaBrowser/.test(ua)) {
		return { ...android, schritte: [schritt('Rechts neben der Adresszeile auf {icon} tippen.', punkte), schritt('„Installieren und Verknüpfung erstellen“ tippen.'), schritt('„Installieren“ tippen.')] };
	}
	return null;
}

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
