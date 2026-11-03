<script>
	import { zustand } from '../../lib/zustand.svelte.js';
	import { gehe } from '../../lib/weg.svelte.js';
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';
	import Button from '../ui/Button.svelte';
	import RaumWahl from '../RaumWahl.svelte';
	import RaeumeHinzufuegen from './RaeumeHinzufuegen.svelte';
	import RaumBearbeiten from './RaumBearbeiten.svelte';

	/**
	 * Die Räume wie beim Melden (RaumWahl): suchen oder erst den Bereich, dann den Raum. Ein Tipp öffnet
	 * ihn zum Umbenennen oder Löschen. Ohne Räume steht das Formular zum Anlegen gleich da; sonst öffnet
	 * es „Räume hinzufügen“ als Blatt, damit die Räume oben stehen.
	 */
	let hinzufuegenOffen = $state(false);
	let bearbeitet = $state(null);
</script>

<main class="mx-auto flex max-w-3xl flex-col gap-4 px-4 pt-6 pb-12">
	{#if zustand.raeume.length === 0}
		<section class="rounded-xl bg-surface-container-low p-4">
			<h2 class="mb-2 text-title-medium text-on-surface">Räume anlegen</h2>
			<RaeumeHinzufuegen />
		</section>
	{:else}
		<div class="flex flex-wrap gap-2">
			<Button variant="tonal" onclick={() => (hinzufuegenOffen = true)}><Icon name="add" groesse={20} />Räume hinzufügen</Button>
			<Button variant="secondary" onclick={() => gehe('/raeume/qr')}><Icon name="qr_code_2" groesse={20} />QR-Aufkleber drucken</Button>
		</div>

		<div><RaumWahl raeume={zustand.raeume} onwahl={(r) => (bearbeitet = r)} /></div>
	{/if}
</main>

<Blatt bind:offen={hinzufuegenOffen} titel="Räume hinzufügen">
	<div class="px-6 pb-6"><RaeumeHinzufuegen fertig={() => (hinzufuegenOffen = false)} /></div>
</Blatt>
<RaumBearbeiten bind:raum={bearbeitet} />
