<script>
	import { zurueck } from '../lib/weg.svelte.js';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Reiter from '../components/ui/Reiter.svelte';
	import RaeumeVerwalten from '../components/einstellungen/RaeumeVerwalten.svelte';
	import ArtenSchalten from '../components/einstellungen/ArtenSchalten.svelte';
	import PersonenVerwalten from '../components/einstellungen/PersonenVerwalten.svelte';
	import VersandPostfach from '../components/einstellungen/VersandPostfach.svelte';

	/** Einstellungen für alle, die alles dürfen: Räume mit QR-Aufklebern, Kacheln des Formulars, Personen, Postfach für den Versand. */
	let reiter = $state('raeume');
</script>

<KopfLeiste titel="Einstellungen" zurueck={() => zurueck('/')} />

<div class="mx-auto max-w-3xl px-4">
	<Reiter
		etikett="Einstellungen"
		reiter={[
			{ id: 'raeume', label: 'Räume' },
			{ id: 'arten', label: 'Arten' },
			{ id: 'personen', label: 'Personen' },
			{ id: 'mail', label: 'Mail' }
		]}
		aktiv={reiter}
		onwahl={(id) => (reiter = id)}
	/>
</div>

{#if reiter === 'raeume'}
	<RaeumeVerwalten />
{:else if reiter === 'arten'}
	<ArtenSchalten />
{:else if reiter === 'personen'}
	<PersonenVerwalten />
{:else}
	<VersandPostfach />
{/if}
