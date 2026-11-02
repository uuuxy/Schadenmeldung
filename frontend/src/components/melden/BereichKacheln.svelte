<script>
	import { WEITERE } from '../../lib/raumgruppen.js';

	/**
	 * Die Bereiche als Kacheln: erst der Bereich, dann der Raum, zwei Tipps statt einer langen Liste.
	 * Kurze Namen („C“) stehen zu viert in einer Reihe wie Tasten, lange zu zweit. Die letzte Kachel
	 * füllt ihre Reihe, damit das Raster aufgeht wie bei den Arten; „Weitere Räume“ nennt, was darin
	 * liegt, sonst sucht niemand die Aula dort.
	 */
	let { gruppen, onwahl } = $props();

	const kurz = $derived(gruppen.every((g) => g.titel === WEITERE || g.titel.length <= 3));
	const spalten = $derived(kurz ? 4 : 2);
	// Spalten, die der letzten Kachel in ihrer Reihe bleiben; die Klassen stehen ausgeschrieben für Tailwind.
	const breiteLetzte = $derived(['', 'col-span-2', 'col-span-3', 'col-span-4'][spalten - ((gruppen.length - 1) % spalten) - 1]);
	const inhalt = (g) =>
		g.titel === WEITERE ? g.raeume.slice(0, 4).map((r) => r.name).join(' · ') + (g.raeume.length > 4 ? ' …' : '') : g.raeume.length === 1 ? '1 Raum' : `${g.raeume.length} Räume`;
</script>

<div class={['grid gap-2', kurz ? 'grid-cols-4' : 'grid-cols-2']} role="group" aria-label="Bereich">
	{#each gruppen as g, i (g.titel)}
		{@const breit = i === gruppen.length - 1 && breiteLetzte !== ''}
		{@const mittig = g.titel.length <= 3 && !breit}
		<button
			class={[
				'flex min-h-16 min-w-0 flex-col justify-center rounded-xl bg-surface-container-low px-3 py-2 text-on-surface transition-transform duration-200 active:scale-[0.97]',
				mittig ? 'items-center text-center' : 'items-start text-left',
				i === gruppen.length - 1 && breiteLetzte
			]}
			aria-label="{g.titel}: {inhalt(g)}"
			onclick={() => onwahl(g)}
		>
			<span class={[mittig ? 'text-title-large' : 'text-title-medium', 'max-w-full truncate']}>{g.titel}</span>
			<span class="line-clamp-1 max-w-full text-body-small text-on-surface-variant">{inhalt(g)}</span>
		</button>
	{/each}
</div>
