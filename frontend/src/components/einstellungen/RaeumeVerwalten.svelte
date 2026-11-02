<script>
	import { zustand } from '../../lib/zustand.svelte.js';
	import { gehe } from '../../lib/weg.svelte.js';
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';
	import Button from '../ui/Button.svelte';
	import RaeumeHinzufuegen from './RaeumeHinzufuegen.svelte';
	import RaumBearbeiten from './RaumBearbeiten.svelte';

	/**
	 * Die Raumliste nach Bereichen, als segmentierte M3-Liste. Ein Tipp auf einen Raum öffnet ihn zum
	 * Umbenennen oder Löschen. Ohne Räume steht das Formular zum Anlegen gleich da; sonst öffnet es
	 * „Räume hinzufügen“ als Blatt, damit die Liste oben steht.
	 */
	let hinzufuegenOffen = $state(false);
	let bearbeitet = $state(null);
	const gruppen = $derived.by(() => {
		const bereiche = new Map();
		for (const r of zustand.raeume) {
			const b = r.bereich || 'Ohne Bereich';
			if (!bereiche.has(b)) bereiche.set(b, []);
			bereiche.get(b).push(r);
		}
		return [...bereiche].map(([bereich, raeume]) => ({ bereich, raeume }));
	});
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

		{#each gruppen as g (g.bereich)}
			<section>
				<h3 class="px-1 pt-2 pb-2 text-title-small text-primary">{g.bereich}</h3>
				<ul class="flex flex-col gap-0.5">
					{#each g.raeume as r (r.id)}
						<li class="rounded-xs bg-surface-container-low first:rounded-t-lg last:rounded-b-lg">
							<button class="flex min-h-14 w-full items-center gap-4 rounded-[inherit] px-4 text-left" onclick={() => (bearbeitet = r)}>
								<Icon name="meeting_room" class="text-on-surface-variant" />
								<span class="min-w-0 flex-1 truncate text-body-large text-on-surface">{r.name}</span>
								<Icon name="chevron_right" class="text-on-surface-variant" />
							</button>
						</li>
					{/each}
				</ul>
			</section>
		{/each}
	{/if}
</main>

<Blatt bind:offen={hinzufuegenOffen} titel="Räume hinzufügen">
	<div class="px-6 pb-6"><RaeumeHinzufuegen fertig={() => (hinzufuegenOffen = false)} /></div>
</Blatt>
<RaumBearbeiten bind:raum={bearbeitet} />
