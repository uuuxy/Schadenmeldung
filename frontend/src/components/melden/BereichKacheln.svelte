<script>
	import Kachel from './Kachel.svelte';
	import { WEITERE } from '../../lib/raumgruppen.js';

	/**
	 * Die Bereiche als Kacheln: erst der Bereich, dann der Raum, zwei Tipps statt einer langen Liste.
	 * Alle Kacheln sind gleich groß, auch „Weitere Räume“ und, beim Melden, „Anderer Ort“ am Ende (M3:
	 * das Standard-Raster zeigt Karten gleich groß); die letzte Reihe darf kürzer sein. Kurze Namen („C“)
	 * stehen wie Tasten eines Telefons zu dritt, ab sm zu viert; die beiden tragen dort ein Symbol statt
	 * des Buchstabens, wie Stern und Raute, damit ihr Name in dieselbe Kachel passt. Lange Namen stehen
	 * zu zweit, und „Weitere Räume“ nennt darunter, was darin liegt, sonst sucht niemand die Aula dort.
	 */
	let { gruppen, onwahl, anderOrt = null } = $props();

	const bereiche = $derived(gruppen.filter((g) => g.titel !== WEITERE));
	const weitere = $derived(gruppen.find((g) => g.titel === WEITERE));
	const kurz = $derived(bereiche.every((g) => g.titel.length <= 3));
	const anzahl = (g) => (g.raeume.length === 1 ? '1 Raum' : `${g.raeume.length} Räume`);
	const namen = (g) => g.raeume.slice(0, 4).map((r) => r.name).join(' · ') + (g.raeume.length > 4 ? ' …' : '');
</script>

<div class={['grid gap-3', kurz ? 'grid-cols-3 sm:grid-cols-4' : 'grid-cols-2']} role="group" aria-label="Bereich">
	{#each bereiche as g (g.titel)}
		<Kachel {kurz} oben={g.titel} unten={anzahl(g)} etikett="{g.titel}: {anzahl(g)}" onclick={() => onwahl(g)} />
	{/each}
	{#if weitere}<Kachel {kurz} oben={WEITERE} unten={namen(weitere)} symbol="more_horiz" etikett="{WEITERE}: {namen(weitere)}" onclick={() => onwahl(weitere)} />{/if}
	{#if anderOrt}<Kachel {kurz} oben="Anderer Ort" unten="Schulhof, Flur …" symbol="edit_location_alt" etikett="Anderer Ort: Schulhof, Flur …" onclick={anderOrt} />{/if}
</div>
