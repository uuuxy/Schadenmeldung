<script>
	import Icon from '../Icon.svelte';
	import { stati, zwischenstati } from '../../lib/status.js';

	/**
	 * Der Zwischenstand als verbundene Knopfgruppe (M3 Expressive „connected button group“, Einfach-
	 * auswahl): drei Knöpfe über die ganze Breite, der gewählte wird zur Pille in secondary-container
	 * und trägt ein Häkchen. Ein zweiter Tipp nimmt ihn zurück auf „Offen“. Was nach dem Tipp
	 * geschieht (speichern, zurück zur Übersicht), entscheidet die Seite über onwahl.
	 */
	let { meldung, onwahl } = $props();
</script>

<div class="grid grid-cols-3 gap-0.5" role="group" aria-label="Stand">
	{#each zwischenstati as s, i (s)}
		{@const gewaehlt = meldung.status === s}
		<button
			class={[
				'treffer flex h-10 min-w-0 items-center justify-center gap-1.5 px-2 text-label-large transition-[background-color,color,border-radius] duration-300',
				gewaehlt
					? 'rounded-full bg-secondary-container text-on-secondary-container'
					: ['bg-surface-container-highest text-on-surface-variant', i === 0 ? 'rounded-l-full rounded-r-sm' : i === 2 ? 'rounded-l-sm rounded-r-full' : 'rounded-sm']
			]}
			aria-pressed={gewaehlt}
			onclick={() => onwahl(s)}
		>
			<Icon name={gewaehlt ? 'check' : stati[s].icon} groesse={18} /><span class="truncate">{stati[s].knapp}</span>
		</button>
	{/each}
</div>
