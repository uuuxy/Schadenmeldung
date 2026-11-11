<script>
	import Icon from '../Icon.svelte';
	import StatusChip from '../ui/StatusChip.svelte';
	import { zustand, darfBearbeiten } from '../../lib/zustand.svelte.js';
	import { abhaken } from '../../lib/meldung.js';
	import { kategorieNach } from '../../lib/kategorien.js';
	import { wann } from '../../lib/zeit.js';
	import { gehe } from '../../lib/weg.svelte.js';
	import { rueckmeldung } from '../../lib/haptik.js';

	/**
	 * Eine Meldung als Eintrag einer M3-Liste (expressive, segmentiert): Ein Tipp öffnet die Meldung,
	 * hinten liegt als einzige Zusatzaktion das Häkchen „Erledigt“ („multi-action list“). Alles
	 * Weitere – Stand, Vermerk, Löschen – steht auf der Seite der Meldung, nicht an jeder Zeile.
	 * Offen ist der Normalfall und trägt keinen Chip; einen Chip gibt es erst ab „In Arbeit“. Der
	 * Bereich steht auf der Seite der Meldung und im Filter; hier bekommt der Ort die ganze Zeile.
	 *
	 * Auswahl (M3 selection, wie in Gmail): Langes Drücken beginnt sie (onauswahl). Solange sie läuft
	 * (auswahl: die gewählten IDs), ist die ganze Zeile die Auswahl: Ein Tipp wählt dazu oder ab, vorn
	 * steht dann das Häkchen statt der Art, und „Erledigt“ fehlt, denn in der Auswahl hat eine Zeile
	 * keine zweite Aktion (M3: „multi-select list items can't have secondary nested actions“).
	 */
	let { meldung: m, auswahl = null, onauswahl } = $props();
	const art = $derived(kategorieNach(m.kategorie));
	const erledigt = $derived(m.status === 'erledigt');
	const gewaehlt = $derived(!!auswahl?.includes(m.id));
	const zeit = $derived(erledigt && m.erledigtAm ? `erledigt ${wann(m.erledigtAm)}` : wann(m.gemeldetAm));
	// Im Funkloch geänderter Stand, der noch im Ausgang liegt: sichtbar, damit niemand nachtippt.
	const wartet = $derived(zustand.wartend.some((w) => w.meldungId === m.id));

	// Langes Drücken (eine halbe Sekunde wie bei Android) beginnt die Auswahl; wer den Finger bewegt,
	// scrollt und bricht es ab. Das Loslassen danach ist kein Tipp.
	let gedrueckt = null;
	let lang = false;
	function druecke(e) {
		if (auswahl || e.button !== 0) return;
		lang = false;
		const uhr = setTimeout(() => {
			gedrueckt = null;
			lang = true;
			rueckmeldung();
			onauswahl(m);
		}, 500);
		gedrueckt = { x: e.clientX, y: e.clientY, uhr };
	}
	function lasse() {
		if (gedrueckt) clearTimeout(gedrueckt.uhr);
		gedrueckt = null;
	}
	function tippe() {
		if (lang) lang = false;
		else if (auswahl) onauswahl(m);
		else gehe(`/meldung/${m.id}`);
	}
</script>

<div class="flex items-center gap-1">
	<button
		class="flex min-w-0 flex-1 items-start gap-4 rounded-[inherit] py-3 pr-2 pl-4 text-left select-none [-webkit-touch-callout:none]"
		aria-pressed={auswahl ? gewaehlt : undefined}
		onclick={tippe}
		onpointerdown={druecke}
		onpointermove={(e) => gedrueckt && Math.hypot(e.clientX - gedrueckt.x, e.clientY - gedrueckt.y) > 10 && lasse()}
		onpointerup={lasse}
		onpointercancel={lasse}
		onpointerleave={lasse}
		oncontextmenu={(e) => (gedrueckt || lang) && e.preventDefault()}
	>
		<span class={['mt-0.5 grid size-10 shrink-0 place-items-center rounded-full transition-colors', gewaehlt ? 'bg-primary text-on-primary' : erledigt ? 'bg-success-container text-on-success-container' : art.farbe.flaeche]}>
			<Icon name={gewaehlt || erledigt ? 'check' : art.icon} groesse={22} class={gewaehlt || erledigt ? '' : art.farbe.symbol} />
		</span>
		<span class="min-w-0 flex-1">
			<span class={['block truncate text-title-medium', gewaehlt ? 'text-on-secondary-container' : 'text-on-surface']}>{m.ort}</span>
			<span class={['block truncate text-body-large', gewaehlt ? 'text-on-secondary-container' : 'text-on-surface']}>{m.was}</span>
			<span class="mt-1 flex items-center gap-2 text-body-small text-on-surface-variant">
				{#if m.dringend && !erledigt}<StatusChip dringend />{/if}
				{#if !erledigt && m.status !== 'offen'}<StatusChip status={m.status} knapp />{/if}
				<span class="truncate">{zeit}</span>
				{#if m.foto}<Icon name="photo_camera" groesse={16} class="shrink-0" />{/if}
				{#if wartet}<span class="inline-flex shrink-0 items-center gap-1"><Icon name="schedule_send" groesse={16} />wartet auf Netz</span>{/if}
			</span>
			{#if m.antwort}<span class="mt-1 block truncate text-body-small text-on-surface-variant">Vermerk: {m.antwort}</span>{/if}
		</span>
	</button>
	{#if darfBearbeiten() && !erledigt && !auswahl}
		<button class="icon-btn mr-2 shrink-0 border border-outline-variant text-on-surface-variant" aria-label="Erledigt: {m.ort}, {m.was}" title="Erledigt" onclick={() => abhaken(m)}>
			<Icon name="check" groesse={22} />
		</button>
	{/if}
</div>
