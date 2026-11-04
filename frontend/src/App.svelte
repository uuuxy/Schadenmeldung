<script>
	import { onMount } from 'svelte';
	import { zustand, starte, auffrischen, darfSehen, darfBearbeiten } from './lib/zustand.svelte.js';
	import { weg } from './lib/weg.svelte.js';
	import { gesehen } from './lib/mitteilungen.js';
	import Anmelden from './views/Anmelden.svelte';
	import Melden from './views/Melden.svelte';
	import Uebersicht from './views/Uebersicht.svelte';
	import Einstellungen from './views/Einstellungen.svelte';
	import MeldungDetail from './views/MeldungDetail.svelte';
	import Aktion from './views/Aktion.svelte';
	import Datenschutz from './views/Datenschutz.svelte';
	import Snackbar from './components/Snackbar.svelte';
	import Rueckfrage from './components/Rueckfrage.svelte';
	import Ladekreis from './components/Ladekreis.svelte';

	// Nach dem Start gleich nachreichen, was im Funkloch liegen geblieben ist.
	onMount(() => starte().then(auffrischen));

	// Wer wohin kommt: Das Kollegium sieht nur „Schaden melden“. Alle mit Übersicht landen bei den
	// Meldungen, öffnen einzelne und drucken QR-Aufkleber; wer alles darf, kommt auch zu den
	// Einstellungen. Der QR-Code (/r/…) führt jeden direkt ins Formular, mit dem Raum schon gewählt.
	const ansicht = $derived.by(() => {
		const p = weg.pfad;
		if (p.startsWith('/aktion/')) return 'aktion';
		if (p === '/datenschutz') return 'datenschutz';
		if (p.startsWith('/r/') || p === '/melden') return 'melden';
		if (p.startsWith('/meldung/') && darfSehen()) return 'meldung';
		if ((p === '/einstellungen' || p === '/verwaltung') && darfBearbeiten()) return 'einstellungen';
		if (p === '/raeume/qr' && darfSehen()) return 'qr';
		return darfSehen() ? 'uebersicht' : 'melden';
	});
	const raumCode = $derived(weg.pfad.startsWith('/r/') ? decodeURIComponent(weg.pfad.slice(3)) : '');

	// Live für alle mit Übersicht: Der Server meldet jede Änderung über /api/ereignisse, die App holt
	// dann den Stand. Nur solange sie sichtbar ist, denn ein offener Strom hält das Funkmodul wach.
	$effect(() => {
		if (!zustand.benutzer || !darfSehen()) return;
		let strom = null;
		let warte = 0;
		let ersteVerbindung = true;
		// Mehrere Änderungen kurz hintereinander holen den Stand nur einmal.
		const nachladen = () => {
			clearTimeout(warte);
			warte = setTimeout(auffrischen, 150);
		};
		const oeffne = () => {
			if (strom || document.visibilityState !== 'visible') return;
			strom = new EventSource('/api/ereignisse');
			strom.addEventListener('meldungen', nachladen);
			// Nach einer Lücke (Funkloch, Hintergrund) holt die neue Verbindung nach, was dazwischen kam.
			strom.addEventListener('open', () => (ersteVerbindung ? (ersteVerbindung = false) : nachladen()));
		};
		const schliesse = () => {
			strom?.close();
			strom = null;
		};
		const sichtbarkeit = () => (document.visibilityState === 'visible' ? (oeffne(), nachladen()) : schliesse());
		oeffne();
		document.addEventListener('visibilitychange', sichtbarkeit);
		return () => {
			schliesse();
			clearTimeout(warte);
			document.removeEventListener('visibilitychange', sichtbarkeit);
		};
	});

	// Wer die App öffnet, hat die neuen Meldungen gesehen: Zahl am App-Symbol und Mitteilungen
	// verschwinden, und das Abo meldet sich beim Server (lib/mitteilungen.js).
	$effect(() => {
		if (!zustand.benutzer || !darfSehen()) return;
		const sichtbar = () => document.visibilityState === 'visible' && gesehen().catch(() => {});
		sichtbar();
		document.addEventListener('visibilitychange', sichtbar);
		return () => document.removeEventListener('visibilitychange', sichtbar);
	});

	// Ohne Strom (Kollegium, oder ein Proxy schluckt ihn) kommen neue Meldungen trotzdem: alle 30
	// Sekunden, sobald die App wieder sichtbar wird und sobald das Netz zurück ist. Dabei geht auch der
	// Ausgang raus.
	$effect(() => {
		if (!zustand.benutzer) return;
		const sichtbar = () => document.visibilityState === 'visible' && auffrischen();
		const takt = setInterval(sichtbar, 30_000);
		document.addEventListener('visibilitychange', sichtbar);
		window.addEventListener('online', auffrischen);
		return () => {
			clearInterval(takt);
			document.removeEventListener('visibilitychange', sichtbar);
			window.removeEventListener('online', auffrischen);
		};
	});
</script>

{#if ansicht === 'aktion'}
	<!-- Die Knöpfe aus der Mail wirken ohne Anmeldung: Der Link selbst berechtigt. -->
	<Aktion token={decodeURIComponent(weg.pfad.slice(8))} />
{:else if !zustand.bereit}
	<div class="grid min-h-dvh place-items-center text-primary"><Ladekreis groesse={40} /></div>
{:else if ansicht === 'datenschutz'}
	<Datenschutz />
{:else if !zustand.benutzer}
	<Anmelden />
{:else if ansicht === 'qr'}
	<!-- Der QR-Erzeuger wird nur für die Verwaltung geladen und bleibt aus dem Startpaket. -->
	{#await import('./views/QrBogen.svelte') then modul}
		{@const Bogen = modul.default}
		<Bogen />
	{/await}
{:else if ansicht === 'einstellungen'}
	<Einstellungen />
{:else if ansicht === 'meldung'}
	{#key weg.pfad}<MeldungDetail id={Number(weg.pfad.slice(9))} />{/key}
{:else if ansicht === 'uebersicht'}
	<Uebersicht />
{:else}
	{#key raumCode}<Melden {raumCode} />{/key}
{/if}
<Snackbar />
<Rueckfrage />
