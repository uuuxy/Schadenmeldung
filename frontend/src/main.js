import './app.css';
import './lib/darstellung.svelte.js';
// Früh laden: Chrome meldet gleich beim Start, dass es die App installieren kann.
import './lib/installieren.svelte.js';
import './lib/fassung.js';
import { mount } from 'svelte';
import App from './App.svelte';

mount(App, { target: document.getElementById('app') });
// Safari am iPhone zeigt :active (Druck-Schicht, Formwechsel beim Drücken) nur, wenn jemand auf
// Berührungen hört; ohne diesen leeren Lauscher sähe man beim Antippen nichts.
document.addEventListener('touchstart', () => {}, { passive: true });

// Der Service Worker lässt die App sofort starten, auch im Funkloch. Er braucht https (oder localhost).
if ('serviceWorker' in navigator && window.isSecureContext) {
	navigator.serviceWorker.register('/sw.js').catch(() => {});
}
