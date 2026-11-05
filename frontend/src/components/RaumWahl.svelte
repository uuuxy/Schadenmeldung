<script>
	import { raumGruppen } from '../lib/raumgruppen.js';
	import Suchpille from './ui/Suchpille.svelte';
	import BereichKacheln from './melden/BereichKacheln.svelte';
	import ZurueckZeile from './ZurueckZeile.svelte';
	import Kachel from './melden/Kachel.svelte';
	import RaumRaster from './melden/RaumRaster.svelte';

	/**
	 * Räume finden wie die Raumsuche in Google Kalender: oben suchen („204“), sonst die Bereiche als
	 * Kacheln und nach einem Tipp nur deren Räume, nach Stockwerken. Wenige Räume stehen gleich da.
	 * Melden und Raumverwaltung teilen sich diesen Weg, damit Räume überall gleich aussehen.
	 * vorschlaege steht unter der Suche, solange nichts eingegeben ist (beim Melden „Zuletzt“);
	 * ohneTreffer unter den Treffern, wenn kein Raum genau so heißt. anderOrt (nur beim Melden) führt zur
	 * freien Ortsangabe: als letzte Kachel bei den Bereichen, bei wenigen Räumen als Kachel unter ihnen;
	 * bei Suche und in einem Bereich nicht, dort gibt es „… als Ort nehmen“ und den Pfeil zurück.
	 */
	let { raeume, onwahl, suche = $bindable(''), vorschlaege, ohneTreffer, anderOrt = null } = $props();

	// Bis zu so vielen Räumen genügt ein Blick; darüber erst der Bereich, dann der Raum.
	const AUF_EINEN_BLICK = 15;
	let gruppenTitel = $state(null);
	const gruppen = $derived(raumGruppen(raeume));
	// Wenige Räume stehen gruppenweise untereinander; alle teilen eine Kachelbreite.
	const alleKurz = $derived(raeume.every((r) => r.name.length <= 5));
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
	<ZurueckZeile titel={gruppe.titel} etikett="Zurück zu allen Bereichen" onzurueck={() => (gruppenTitel = null)} />
	<div class="mt-1"><RaumRaster raeume={gruppe.raeume} stockwerke={gruppe.stockwerke} onwahl={waehle} /></div>
{:else}
	{@render vorschlaege?.()}
	{#if stufenweise}
		<h3 class="mt-5 mb-2 text-label-large text-on-surface-variant">Bereich</h3>
		<BereichKacheln {gruppen} {anderOrt} onwahl={(g) => (gruppenTitel = g.titel)} />
	{:else}
		{#each gruppen as g (g.titel)}
			{#if gruppen.length > 1}<h3 class="mt-5 mb-2 text-title-small text-primary">{g.titel}</h3>{/if}
			<div class={gruppen.length > 1 ? '' : 'mt-4'}><RaumRaster raeume={g.raeume} stockwerke={g.stockwerke} kurz={alleKurz} onwahl={waehle} /></div>
		{/each}
		{#if anderOrt}
			<div class={['mt-4 grid gap-3', alleKurz ? 'grid-cols-3 sm:grid-cols-4' : 'grid-cols-2']}>
				<Kachel kurz={alleKurz} oben="Anderer Ort" unten="Schulhof, Flur …" symbol="edit_location_alt" etikett="Anderer Ort: Schulhof, Flur …" onclick={anderOrt} />
			</div>
		{/if}
	{/if}
{/if}
