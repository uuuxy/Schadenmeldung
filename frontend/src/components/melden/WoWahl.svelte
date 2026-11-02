<script>
	import { zustand, darfBearbeiten } from '../../lib/zustand.svelte.js';
	import { rueckmeldung } from '../../lib/haptik.js';
	import { zuletztGewaehlt, merkeRaum } from '../../lib/zuletzt.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import Feld from '../ui/Feld.svelte';
	import FilterChips from '../ui/FilterChips.svelte';
	import Suchpille from '../ui/Suchpille.svelte';
	import HierGemeldet from './HierGemeldet.svelte';

	/**
	 * Wo? Mit QR-Code ist der Raum schon gewählt. Sonst: ein Tipp auf „Zuletzt“, zwei Tipps über
	 * Bereich → Raum, oder kurz „204“ suchen. Für alles ohne Raum gibt es „Anderer Ort“. Gewählt
	 * schrumpft die Frage auf eine Zeile mit „Ändern“.
	 */
	let { raum = $bindable(), ortFrei = $bindable(), freierOrt = $bindable(), codeUnbekannt = false, ongewaehlt } = $props();

	const OHNE_BEREICH = 'Weitere Räume';
	let suche = $state('');
	let bereich = $state(undefined);
	let zuletzt = $state(zuletztGewaehlt());

	const bereichVon = (r) => r.bereich || OHNE_BEREICH;
	const zuletztRaeume = $derived(zuletzt.map((id) => zustand.raeume.find((r) => r.id === id)).filter(Boolean));
	const bereiche = $derived([...new Set(zustand.raeume.map(bereichVon))]);
	// Vorgewählt ist der Bereich des zuletzt gewählten Raums: Meist meldet man aus derselben Ecke.
	// Wer den Chip abwählt (null), sieht alle Räume nach Bereichen.
	const aktiverBereich = $derived(
		bereich !== undefined ? bereich : zuletztRaeume[0] ? bereichVon(zuletztRaeume[0]) : bereiche.length > 1 ? bereiche[0] : null
	);
	const gruppen = $derived(
		bereiche
			.filter((b) => aktiverBereich === null || b === aktiverBereich)
			.map((b) => ({ bereich: b, raeume: zustand.raeume.filter((r) => bereichVon(r) === b) }))
	);
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

{#snippet raumKnopf(r, mitBereich)}
	<button
		class="flex min-h-12 flex-col items-center justify-center rounded-md border border-outline-variant bg-surface-container-lowest px-2 py-2 text-center text-on-surface transition-transform duration-200 active:scale-[0.97]"
		onclick={() => waehleRaum(r)}
	>
		<span class="text-label-large [overflow-wrap:anywhere]">{r.name}</span>
		{#if mitBereich && r.bereich}<span class="mt-0.5 line-clamp-1 text-body-small text-on-surface-variant">{r.bereich}</span>{/if}
	</button>
{/snippet}

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
		<div class="mt-4 grid grid-cols-[repeat(auto-fill,minmax(6.5rem,1fr))] gap-2">
			{#each treffer as r (r.id)}{@render raumKnopf(r, true)}{/each}
		</div>
		{#if !treffer.some((r) => r.name.toLowerCase() === begriff)}
			<Button variant="ghost" class="mt-2 -ml-4" onclick={ortAusSuche}><Icon name="location_on" groesse={18} />„{suche.trim()}“ als Ort nehmen</Button>
		{/if}
	{:else}
		{#if zuletztRaeume.length}
			<p class="mt-5 mb-2 text-label-large text-on-surface-variant">Zuletzt</p>
			<FilterChips etikett="Zuletzt gewählte Räume" optionen={zuletztRaeume.map((r) => ({ wert: String(r.id), text: r.name, icon: 'history' }))} wert={null} onwahl={(id) => waehleRaum(zuletztRaeume.find((r) => String(r.id) === id))} />
		{/if}
		{#if bereiche.length > 1}
			<p class="mt-5 mb-2 text-label-large text-on-surface-variant">Bereich</p>
			<FilterChips etikett="Bereich" optionen={bereiche.map((b) => ({ wert: b, text: b }))} wert={aktiverBereich} onwahl={(b) => (bereich = b)} />
		{/if}
		{#each gruppen as g (g.bereich)}
			{#if aktiverBereich === null && bereiche.length > 1}<p class="mt-5 mb-2 text-title-small text-primary">{g.bereich}</p>{/if}
			<div class="mt-3 grid grid-cols-[repeat(auto-fill,minmax(6.5rem,1fr))] gap-2">
				{#each g.raeume as r (r.id)}{@render raumKnopf(r, false)}{/each}
			</div>
		{/each}
	{/if}
	<Button variant="ghost" class="mt-3 -ml-4" onclick={() => (freierOrt = true)}><Icon name="edit_note" groesse={18} />Anderer Ort (Schulhof, Flur …)</Button>
{/if}
