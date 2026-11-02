<script>
	import Icon from '../Icon.svelte';
	import { stati, zwischenstati } from '../../lib/status.js';
	import { zwischenstand } from '../../lib/meldung.js';

	/**
	 * Der Zwischenstand als verbundene Knopfgruppe (M3 Expressive „connected button group“, Einfach-
	 * auswahl): drei Knöpfe über die ganze Breite, der gewählte wird zur Pille in secondary-container
	 * und trägt ein Häkchen. Ein zweiter Tipp nimmt ihn zurück auf „Offen“.
	 */
	let { meldung } = $props();
	let laeuft = $state(false);

	async function setze(s) {
		laeuft = true;
		await zwischenstand(meldung, s);
		laeuft = false;
	}
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
			disabled={laeuft}
			onclick={() => setze(s)}
		>
			<Icon name={gewaehlt ? 'check' : stati[s].icon} groesse={18} /><span class="truncate">{stati[s].knapp}</span>
		</button>
	{/each}
</div>
