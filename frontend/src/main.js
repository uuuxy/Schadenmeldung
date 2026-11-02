import './app.css';
import './lib/darstellung.svelte.js';
import { mount } from 'svelte';
import App from './App.svelte';

mount(App, { target: document.getElementById('app') });

// Der Service Worker lässt die App sofort starten, auch im Funkloch. Er braucht https (oder localhost).
if ('serviceWorker' in navigator && window.isSecureContext) {
	navigator.serviceWorker.register('/sw.js').catch(() => {});
}
