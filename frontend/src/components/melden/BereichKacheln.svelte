<script>
	import { WEITERE } from '../../lib/raumgruppen.js';

	/**
	 * Die Bereiche als Kacheln: erst der Bereich, dann der Raum, zwei Tipps statt einer langen Liste.
	 * Kurze Namen („C“) stehen zu viert in einer Reihe wie Tasten, lange zu zweit. „Weitere Räume“ und,
	 * beim Melden, „Anderer Ort“ teilen sich eine eigene letzte Reihe zu gleichen Teilen: zwei gleich
	 * große Kacheln, statt dass eine den Rest der Reihe füllt und die andere als kleiner Link darunter
	 * steht. „Weitere Räume“ nennt, was darin liegt, sonst sucht niemand die Aula dort.
	 */
	let { gruppen, onwahl, anderOrt = null } = $props();

	const bereiche = $derived(gruppen.filter((g) => g.titel !== WEITERE));
	const weitere = $derived(gruppen.find((g) => g.titel === WEITERE));
	const kurz = $derived(bereiche.every((g) => g.titel.length <= 3));
	const inhalt = (g) =>
		g.titel === WEITERE ? g.raeume.slice(0, 4).map((r) => r.name).join(' · ') + (g.raeume.length > 4 ? ' …' : '') : g.raeume.length === 1 ? '1 Raum' : `${g.raeume.length} Räume`;
</script>

{#snippet kachel(titel, text, tun)}
	{@const mittig = titel.length <= 3}
	<button
		class={[
			'flex min-h-16 min-w-0 flex-col justify-center rounded-xl bg-surface-container-low px-3 py-2 text-on-surface transition-transform duration-200 active:scale-[0.97]',
			mittig ? 'items-center text-center' : 'items-start text-left'
		]}
		aria-label="{titel}: {text}"
		onclick={tun}
	>
		<span class={[mittig ? 'text-title-large' : 'text-title-medium', 'max-w-full truncate']}>{titel}</span>
		<span class="line-clamp-1 max-w-full text-body-small text-on-surface-variant">{text}</span>
	</button>
{/snippet}

<div class={['grid gap-2', kurz ? 'grid-cols-4' : 'grid-cols-2']} role="group" aria-label="Bereich">
	{#each bereiche as g (g.titel)}
		{@render kachel(g.titel, inhalt(g), () => onwahl(g))}
	{/each}
	{#if weitere || anderOrt}
		<div class="col-span-full grid auto-cols-fr grid-flow-col gap-2">
			{#if weitere}{@render kachel(weitere.titel, inhalt(weitere), () => onwahl(weitere))}{/if}
			{#if anderOrt}{@render kachel('Anderer Ort', 'Schulhof, Flur …', anderOrt)}{/if}
		</div>
	{/if}
</div>
