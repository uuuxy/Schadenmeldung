<script>
	import { zustand, darfBearbeiten } from '../../lib/zustand.svelte.js';
	import { rueckmeldung } from '../../lib/haptik.js';
	import { zuletztGewaehlt, merkeRaum } from '../../lib/zuletzt.js';
	import { raumGruppen } from '../../lib/raumgruppen.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import Feld from '../ui/Feld.svelte';
	import FilterChips from '../ui/FilterChips.svelte';
	import Suchpille from '../ui/Suchpille.svelte';
	import HierGemeldet from './HierGemeldet.svelte';
	import BereichKacheln from './BereichKacheln.svelte';
	import RaumRaster from './RaumRaster.svelte';

	/**
	 * Wo? Mit QR-Code ist der Raum schon gewählt. Sonst wie die Raumsuche in Google Kalender: oben
	 * suchen („204“), darunter „Zuletzt“ als Vorschläge für einen Tipp, dann die Bereiche als Kacheln
	 * und nach einem Tipp nur deren Räume, nach Stockwerken. Wenige Räume stehen gleich da. Für alles
	 * ohne Raum gibt es „Anderer Ort“. Gewählt schrumpft die Frage auf eine Zeile mit „Ändern“.
	 */
	let { raum = $bindable(), ortFrei = $bindable(), freierOrt = $bindable(), codeUnbekannt = false, ongewaehlt } = $props();

	// Bis zu so vielen Räumen genügt ein Blick; darüber erst der Bereich, dann der Raum.
	const AUF_EINEN_BLICK = 15;
	let suche = $state('');
	let gruppenTitel = $state(null);
	let zuletzt = $state(zuletztGewaehlt());

	const zuletztRaeume = $derived(zuletzt.map((id) => zustand.raeume.find((r) => r.id === id)).filter(Boolean));
	const gruppen = $derived(raumGruppen(zustand.raeume));
	const stufenweise = $derived(zustand.raeume.length > AUF_EINEN_BLICK && gruppen.length > 1);
	const gruppe = $derived(stufenweise ? (gruppen.find((g) => g.titel === gruppenTitel) ?? null) : null);
	const begriff = $derived(suche.trim().toLowerCase());
	const treffer = $derived.by(() => {
		if (!begriff) return [];
		const rang = (r) => (r.name.toLowerCase() === begriff ? 0 : r.name.toLowerCase().startsWith(begriff) ? 1 : 2);
		return zustand.raeume
			.filter((r) => r.name.toLowerCase().includes(begriff) || r.bereich.toLowerCase().includes(begriff))
			.sort((a, b) => rang(a) - rang(b))
			.slice(0, 12);
	});
	const nurFreitext = $derived(zustand.raeume.length === 0 || freierOrt);

	function waehleRaum(r) {
		raum = r;
		suche = '';
		freierOrt = false;
		merkeRaum(r.id);
		zuletzt = zuletztGewaehlt();
		rueckmeldung();
		ongewaehlt?.();
	}

	function ortAusSuche() {
		ortFrei = suche.trim();
		suche = '';
		freierOrt = true;
	}
</script>

{#if codeUnbekannt && !raum}
	<p class="mb-3 rounded-md bg-warning-container px-4 py-3 text-body-medium text-on-warning-container">
		Dieser QR-Code gehört zu keinem Raum mehr. Bitte den Raum auswählen.
	</p>
{/if}

{#if raum}
	<button class="flex w-full items-center gap-4 rounded-xl bg-secondary-container p-4 text-left text-on-secondary-container" onclick={() => (raum = null)}>
		<Icon name="location_on" gefuellt groesse={24} />
		<span class="min-w-0 flex-1">
			<span class="block truncate text-title-medium">{raum.name}</span>
			{#if raum.bereich}<span class="block truncate text-body-medium">{raum.bereich}</span>{/if}
		</span>
		<span class="shrink-0 text-label-large">Ändern</span>
	</button>
	<HierGemeldet raumId={raum.id} />
{:else if nurFreitext}
	<Feld etikett="Wo ist der Schaden?" placeholder="z. B. Schulhof, Treppenhaus Nord" bind:wert={ortFrei} autocomplete="off" enterkeyhint="next" maxlength="80" />
	{#if zustand.raeume.length}
		<Button variant="ghost" class="mt-2 -ml-4" onclick={() => (freierOrt = false)}><Icon name="meeting_room" groesse={18} />Doch einen Raum wählen</Button>
	{:else if darfBearbeiten()}
		<p class="mt-2 text-body-small text-on-surface-variant">Tipp: Unter Einstellungen (Kontomenü) die Raumliste anlegen – dann genügt hier ein Tipp.</p>
	{/if}
{:else}
	<Suchpille bind:wert={suche} platzhalter="Raum suchen, z. B. 204" etikett="Raum suchen" onkeydown={(e) => e.key === 'Enter' && treffer[0] && waehleRaum(treffer[0])} />

	{#if begriff}
		<div class="mt-4"><RaumRaster raeume={treffer} mitBereich onwahl={waehleRaum} /></div>
		{#if !treffer.some((r) => r.name.toLowerCase() === begriff)}
			<Button variant="ghost" class="mt-2 -ml-4" onclick={ortAusSuche}><Icon name="location_on" groesse={18} />„{suche.trim()}“ als Ort nehmen</Button>
		{/if}
	{:else if gruppe}
		<!-- Zurück zu allen Bereichen wie aus einem Untermenü: der Pfeil vorn, daneben, wo man gerade ist. -->
		<button class="mt-3 -ml-2 flex min-h-12 items-center gap-1 rounded-full pr-4 text-on-surface" aria-label="Zurück zu allen Bereichen" onclick={() => (gruppenTitel = null)}>
			<span class="grid size-10 place-items-center"><Icon name="arrow_back" /></span>
			<span class="text-title-medium">{gruppe.titel}</span>
		</button>
		<div class="mt-1"><RaumRaster raeume={gruppe.raeume} stockwerke={gruppe.stockwerke} onwahl={waehleRaum} /></div>
	{:else}
		{#if zuletztRaeume.length}
			<h3 class="mt-5 mb-2 text-label-large text-on-surface-variant">Zuletzt</h3>
			<FilterChips etikett="Zuletzt gewählte Räume" optionen={zuletztRaeume.map((r) => ({ wert: String(r.id), text: r.name, icon: 'history' }))} wert={null} onwahl={(id) => waehleRaum(zuletztRaeume.find((r) => String(r.id) === id))} />
		{/if}
		{#if stufenweise}
			<h3 class="mt-5 mb-2 text-label-large text-on-surface-variant">Bereich</h3>
			<BereichKacheln {gruppen} onwahl={(g) => (gruppenTitel = g.titel)} />
		{:else}
			{#each gruppen as g (g.titel)}
				{#if gruppen.length > 1}<h3 class="mt-5 mb-2 text-title-small text-primary">{g.titel}</h3>{/if}
				<div class={gruppen.length > 1 ? '' : 'mt-4'}><RaumRaster raeume={g.raeume} stockwerke={g.stockwerke} onwahl={waehleRaum} /></div>
			{/each}
		{/if}
	{/if}
	<Button variant="ghost" class="mt-3 -ml-4" onclick={() => (freierOrt = true)}><Icon name="edit_note" groesse={18} />Anderer Ort (Schulhof, Flur …)</Button>
{/if}
