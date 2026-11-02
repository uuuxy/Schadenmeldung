<script>
	import { formate } from '../../lib/etiketten.js';

	/**
	 * Die Druckformate als Kacheln, je mit einem kleinen Blatt, das zeigt, wie die Aufkleber darauf
	 * liegen: So sieht man vor dem Drucken, was herauskommt. Gewählt liegt die Kachel auf
	 * secondary-container wie jede Auswahl.
	 */
	let { wert, onwahl } = $props();
</script>

<div class="grid grid-cols-2 gap-2 sm:grid-cols-4" role="group" aria-label="Format">
	{#each formate as f (f.key)}
		{@const an = f.key === wert}
		<button
			class={['flex min-h-18 items-center gap-3 rounded-xl p-3 text-left transition-colors duration-200', an ? 'bg-secondary-container text-on-secondary-container' : 'bg-surface-container-low text-on-surface']}
			aria-pressed={an}
			onclick={() => onwahl(f.key)}
		>
			<span class="miniblatt" style="--spalten: {f.spalten}; --zeilen: {f.zeilen}" aria-hidden="true">
				{#each { length: f.spalten * f.zeilen } as _, i (i)}<span></span>{/each}
			</span>
			<span class="min-w-0">
				<span class="block text-title-small">{f.name}</span>
				<span class={['block text-body-small', !an && 'text-on-surface-variant']}>{f.kurz}</span>
			</span>
		</button>
	{/each}
</div>

<style>
	/* Das kleine Blatt: weiß wie Papier, die Felder in der Farbe der Schnittlinie. */
	.miniblatt {
		display: grid;
		grid-template-columns: repeat(var(--spalten), 1fr);
		grid-template-rows: repeat(var(--zeilen), 1fr);
		gap: 1px;
		width: 1.75rem;
		aspect-ratio: 210 / 297;
		flex-shrink: 0;
		padding: 2px;
		border-radius: 2px;
		background: var(--color-papier);
		outline: 1px solid var(--color-outline-variant);
	}
	.miniblatt span {
		border-radius: 1px;
		background: var(--color-papier-linie);
	}
</style>
