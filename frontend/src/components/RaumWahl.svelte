<script>
	import { raumGruppen } from '../lib/raumgruppen.js';
	import Icon from './Icon.svelte';
	import Button from './ui/Button.svelte';
	import Suchpille from './ui/Suchpille.svelte';
	import BereichKacheln from './melden/BereichKacheln.svelte';
	import RaumRaster from './melden/RaumRaster.svelte';

	/**
	 * Räume finden wie die Raumsuche in Google Kalender: oben suchen („204“), sonst die Bereiche als
	 * Kacheln und nach einem Tipp nur deren Räume, nach Stockwerken. Wenige Räume stehen gleich da.
	 * Melden und Raumverwaltung teilen sich diesen Weg, damit Räume überall gleich aussehen.
	 * vorschlaege steht unter der Suche, solange nichts eingegeben ist (beim Melden „Zuletzt“);
	 * ohneTreffer unter den Treffern, wenn kein Raum genau so heißt. anderOrt (nur beim Melden) führt zur
	 * freien Ortsangabe: als Kachel bei den Bereichen, bei wenigen Räumen als Link darunter; bei Suche
	 * und in einem Bereich nicht, dort gibt es „… als Ort nehmen“ und den Pfeil zurück.
	 */
	let { raeume, onwahl, suche = $bindable(''), vorschlaege, ohneTreffer, anderOrt = null } = $props();

	// Bis zu so vielen Räumen genügt ein Blick; darüber erst der Bereich, dann der Raum.
	const AUF_EINEN_BLICK = 15;
	let gruppenTitel = $state(null);
	const gruppen = $derived(raumGruppen(raeume));
	const stufenweise = $derived(raeume.length > AUF_EINEN_BLICK && gruppen.length > 1);
	const gruppe = $derived(stufenweise ? (gruppen.find((g) => g.titel === gruppenTitel) ?? null) : null);
	const begriff = $derived(suche.trim().toLowerCase());
	const treffer = $derived.by(() => {
		if (!begriff) return [];
		const rang = (r) => (r.name.toLowerCase() === begriff ? 0 : r.name.toLowerCase().startsWith(begriff) ? 1 : 2);
		return raeume
			.filter((r) => r.name.toLowerCase().includes(begriff) || r.bereich.toLowerCase().includes(begriff))
			.sort((a, b) => rang(a) - rang(b))
			.slice(0, 12);
	});

	function waehle(r) {
		suche = '';
		onwahl(r);
	}
</script>

<Suchpille bind:wert={suche} platzhalter="Raum suchen, z. B. 204" etikett="Raum suchen" onkeydown={(e) => e.key === 'Enter' && treffer[0] && waehle(treffer[0])} />

{#if begriff}
	<div class="mt-4"><RaumRaster raeume={treffer} mitBereich onwahl={waehle} /></div>
	{#if !treffer.some((r) => r.name.toLowerCase() === begriff)}{@render ohneTreffer?.(suche.trim())}{/if}
{:else if gruppe}
	<!-- Zurück zu allen Bereichen wie aus einem Untermenü: der Pfeil vorn, daneben, wo man gerade ist. -->
	<button class="mt-3 -ml-2 flex min-h-12 items-center gap-1 rounded-full pr-4 text-on-surface" aria-label="Zurück zu allen Bereichen" onclick={() => (gruppenTitel = null)}>
		<span class="grid size-10 place-items-center"><Icon name="arrow_back" /></span>
		<span class="text-title-medium">{gruppe.titel}</span>
	</button>
	<div class="mt-1"><RaumRaster raeume={gruppe.raeume} stockwerke={gruppe.stockwerke} onwahl={waehle} /></div>
{:else}
	{@render vorschlaege?.()}
	{#if stufenweise}
		<h3 class="mt-5 mb-2 text-label-large text-on-surface-variant">Bereich</h3>
		<BereichKacheln {gruppen} {anderOrt} onwahl={(g) => (gruppenTitel = g.titel)} />
	{:else}
		{#each gruppen as g (g.titel)}
			{#if gruppen.length > 1}<h3 class="mt-5 mb-2 text-title-small text-primary">{g.titel}</h3>{/if}
			<div class={gruppen.length > 1 ? '' : 'mt-4'}><RaumRaster raeume={g.raeume} stockwerke={g.stockwerke} onwahl={waehle} /></div>
		{/each}
		{#if anderOrt}<Button variant="ghost" class="mt-3 -ml-4" onclick={anderOrt}><Icon name="edit_note" groesse={18} />Anderer Ort (Schulhof, Flur …)</Button>{/if}
	{/if}
{/if}
