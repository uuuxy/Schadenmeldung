<script>
	import { slide } from 'svelte/transition';
	import { kategorieNach } from '../../lib/kategorien.js';
	import Icon from '../Icon.svelte';

	/**
	 * Was ist kaputt? Große Kacheln, jede leuchtet beim Antippen in ihrer eigenen Farbe (kat-* in
	 * rollen.css), wie Geräte-Kacheln in Google Home. Gewählt schrumpft das Raster auf eine Zeile in
	 * dieser Farbe, wie „Wo?“ auf den Raum: Darunter steht sofort die nächste Frage, ohne Wischen.
	 * Abgeschaltete Arten fehlen. Alle Kacheln sind gleich groß, auch „Sonstiges“ (M3: das Standard-Raster
	 * zeigt Karten gleich groß); die letzte Reihe darf kürzer sein.
	 */
	let { arten, kategorie, onwahl } = $props();
	const art = $derived(kategorie ? kategorieNach(kategorie) : null);
</script>

{#if art}
	<div transition:slide={{ duration: 250 }}>
		<button class={['flex w-full items-center gap-4 rounded-xl p-4 text-left', art.farbe.flaeche]} onclick={() => onwahl('')}>
			<Icon name={art.icon} gefuellt groesse={24} class={art.farbe.symbol} />
			<span class="min-w-0 flex-1 truncate text-title-medium">{art.name}</span>
			<span class="shrink-0 text-label-large">Ändern</span>
		</button>
	</div>
{:else}
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-4" role="group" aria-label="Art des Schadens" transition:slide={{ duration: 250 }}>
		{#each arten as k (k.key)}
			<button
				class={[
					'relative flex min-h-28 flex-col justify-between gap-3 rounded-xl bg-surface-container-low p-4 text-left text-on-surface transition-colors duration-200 active:scale-[0.98]'
				]}
				onclick={() => onwahl(k.key)}
			>
				<Icon name={k.icon} groesse={28} class={k.farbe.symbol} />
				<span class="text-title-medium leading-tight">{k.name}</span>
			</button>
		{/each}
	</div>
{/if}
