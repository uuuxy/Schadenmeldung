<script>
	import { api } from '../../lib/api.js';
	import { zustand } from '../../lib/zustand.svelte.js';
	import { zeige } from '../../lib/snackbar.svelte.js';
	import { kategorien } from '../../lib/kategorien.js';
	import Icon from '../Icon.svelte';
	import Switch from '../ui/Switch.svelte';

	/**
	 * Welche Kacheln das Kollegium unter „Was ist kaputt?“ sieht. Ein Schalter je Art (M3: Switch für
	 * Ein/Aus); die übrigen Kacheln rücken im Formular von selbst zusammen. „Sonstiges“ bleibt immer,
	 * als Ausweg für alles andere. Bereits gemeldete Schäden einer abgeschalteten Art bleiben sichtbar.
	 */
	const aus = $derived(zustand.einstellungen.artenAus ?? []);

	async function schalte(art, sichtbar) {
		try {
			const antwort = await api('/api/arten', { methode: 'POST', json: { kategorie: art.key, sichtbar } });
			zustand.einstellungen.artenAus = antwort.artenAus;
			zeige(sichtbar ? `„${art.name}“ steht wieder im Formular` : `„${art.name}“ ist aus dem Formular genommen`);
		} catch (e) {
			zeige(e.message);
		}
	}
</script>

<main class="mx-auto flex max-w-3xl flex-col gap-4 px-4 pt-6 pb-12">
	<p class="text-body-medium text-on-surface-variant">
		Diese Kacheln sieht das Kollegium unter „Was ist kaputt?“. Was abgeschaltet ist, fehlt im Formular; die übrigen Kacheln rücken zusammen.
	</p>
	<ul class="flex flex-col gap-0.5">
		{#each kategorien as art (art.key)}
			{@const fest = art.key === 'sonstiges'}
			<li class="flex min-h-18 items-center gap-4 rounded-xs bg-surface-container-low py-2 pr-4 pl-4 first:rounded-t-lg last:rounded-b-lg">
				<span class={['grid size-10 shrink-0 place-items-center rounded-full', art.farbe.flaeche]}><Icon name={art.icon} groesse={22} class={art.farbe.symbol} /></span>
				<label for={fest ? undefined : `art-${art.key}`} class="min-w-0 flex-1">
					<span class="block text-body-large text-on-surface">{art.name}</span>
					<span class="block truncate text-body-medium text-on-surface-variant">{fest ? 'Bleibt immer, für alles andere' : art.faelle.join(', ')}</span>
				</label>
				{#if fest}
					<span class="shrink-0 text-body-medium text-on-surface-variant">immer</span>
				{:else}
					<Switch id="art-{art.key}" checked={!aus.includes(art.key)} onchange={(an) => schalte(art, an)} />
				{/if}
			</li>
		{/each}
	</ul>
</main>
