<script>
	import { onMount } from 'svelte';
	import { zustand, starte, auffrischen, darfSehen, istAdmin } from './lib/zustand.svelte.js';
	import { weg } from './lib/weg.svelte.js';
	import Anmelden from './views/Anmelden.svelte';
	import Melden from './views/Melden.svelte';
	import Uebersicht from './views/Uebersicht.svelte';
	import Verwaltung from './views/Verwaltung.svelte';
	import Aktion from './views/Aktion.svelte';
	import Datenschutz from './views/Datenschutz.svelte';
	import Snackbar from './components/Snackbar.svelte';
	import Ladekreis from './components/Ladekreis.svelte';

	// Nach dem Start gleich nachreichen, was im Funkloch liegen geblieben ist.
	onMount(() => starte().then(auffrischen));

	// Wer wohin kommt: Das Kollegium sieht nur „Schaden melden“. Hausmeister und Leitung landen bei
	// den Meldungen, die Verwaltung zusätzlich bei Räumen und Personen. Der QR-Code (/r/…) führt
	// jeden direkt ins Formular, mit dem Raum schon gewählt.
	const ansicht = $derived.by(() => {
		const p = weg.pfad;
		if (p.startsWith('/aktion/')) return 'aktion';
		if (p === '/datenschutz') return 'datenschutz';
		if (p.startsWith('/r/') || p === '/melden') return 'melden';
		if (p === '/verwaltung' && istAdmin()) return 'verwaltung';
		if (p === '/raeume/qr' && istAdmin()) return 'qr';
		return darfSehen() ? 'uebersicht' : 'melden';
	});
	const raumCode = $derived(weg.pfad.startsWith('/r/') ? decodeURIComponent(weg.pfad.slice(3)) : '');

	// Neue Meldungen kommen ohne Neuladen: alle 30 Sekunden, sobald die App wieder sichtbar wird und
	// sobald das Netz zurück ist. Dabei geht auch der Ausgang raus.
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
{:else if ansicht === 'verwaltung'}
	<Verwaltung />
{:else if ansicht === 'uebersicht'}
	<Uebersicht />
{:else}
	{#key raumCode}<Melden {raumCode} />{/key}
{/if}
<Snackbar />
