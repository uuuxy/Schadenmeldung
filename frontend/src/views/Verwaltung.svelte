<script>
	import { gehe } from '../lib/weg.svelte.js';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Reiter from '../components/ui/Reiter.svelte';
	import RaeumeVerwalten from '../components/RaeumeVerwalten.svelte';
	import PersonenVerwalten from '../components/PersonenVerwalten.svelte';

	/** Räume & Personen: Räume mit QR-Aufklebern und wer mehr sieht als das Meldeformular. */
	let reiter = $state('raeume');
</script>

<KopfLeiste titel="Räume & Personen" zurueck={() => gehe('/')} />

<div class="mx-auto max-w-3xl px-4">
	<Reiter
		etikett="Räume & Personen"
		reiter={[
			{ id: 'raeume', label: 'Räume' },
			{ id: 'personen', label: 'Personen' }
		]}
		aktiv={reiter}
		onwahl={(id) => (reiter = id)}
	/>
</div>

{#if reiter === 'raeume'}
	<RaeumeVerwalten />
{:else}
	<PersonenVerwalten />
{/if}
