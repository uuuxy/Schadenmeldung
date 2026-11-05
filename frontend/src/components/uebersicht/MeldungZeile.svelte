<script>
	import Icon from '../Icon.svelte';
	import StatusChip from '../ui/StatusChip.svelte';
	import { zustand, darfBearbeiten } from '../../lib/zustand.svelte.js';
	import { abhaken } from '../../lib/meldung.js';
	import { kategorieNach } from '../../lib/kategorien.js';
	import { wann } from '../../lib/zeit.js';
	import { gehe } from '../../lib/weg.svelte.js';

	/**
	 * Eine Meldung als Eintrag einer M3-Liste (expressive, segmentiert): Ein Tipp öffnet die Meldung,
	 * hinten liegt als einzige Zusatzaktion das Häkchen „Erledigt“ („multi-action list“). Alles
	 * Weitere – Stand, Vermerk, Löschen – steht auf der Seite der Meldung, nicht an jeder Zeile.
	 * Offen ist der Normalfall und trägt keinen Chip; einen Chip gibt es erst ab „In Arbeit“. Der
	 * Bereich steht auf der Seite der Meldung und im Filter; hier bekommt der Ort die ganze Zeile.
	 */
	let { meldung: m } = $props();
	const art = $derived(kategorieNach(m.kategorie));
	const erledigt = $derived(m.status === 'erledigt');
	const zeit = $derived(erledigt && m.erledigtAm ? `erledigt ${wann(m.erledigtAm)}` : wann(m.gemeldetAm));
	// Im Funkloch geänderter Stand, der noch im Ausgang liegt: sichtbar, damit niemand nachtippt.
	const wartet = $derived(zustand.wartend.some((w) => w.meldungId === m.id));
</script>

<div class="flex items-center gap-1">
	<button class="flex min-w-0 flex-1 items-start gap-4 rounded-[inherit] py-3 pr-2 pl-4 text-left" onclick={() => gehe(`/meldung/${m.id}`)}>
		<span class={['mt-0.5 grid size-10 shrink-0 place-items-center rounded-full', erledigt ? 'bg-success-container text-on-success-container' : art.farbe.flaeche]}>
			<Icon name={erledigt ? 'check' : art.icon} groesse={22} class={erledigt ? '' : art.farbe.symbol} />
		</span>
		<span class="min-w-0 flex-1">
			<span class="block truncate text-title-medium text-on-surface">{m.ort}</span>
			<span class="block truncate text-body-large text-on-surface">{m.was}</span>
			<span class="mt-1 flex items-center gap-2 text-body-small text-on-surface-variant">
				{#if m.dringend && !erledigt}
					<span class="inline-flex h-6 shrink-0 items-center gap-1 rounded-sm bg-error-container px-2 text-label-medium text-on-error-container"><Icon name="warning" groesse={16} gefuellt />Dringend</span>
				{/if}
				{#if !erledigt && m.status !== 'offen'}<StatusChip status={m.status} knapp />{/if}
				<span class="truncate">{zeit}</span>
				{#if m.foto}<Icon name="photo_camera" groesse={16} class="shrink-0" />{/if}
				{#if wartet}<span class="inline-flex shrink-0 items-center gap-1"><Icon name="schedule_send" groesse={16} />wartet auf Netz</span>{/if}
			</span>
			{#if m.antwort}<span class="mt-1 block truncate text-body-small text-on-surface-variant">Vermerk: {m.antwort}</span>{/if}
		</span>
	</button>
	{#if darfBearbeiten() && !erledigt}
		<button class="icon-btn mr-2 shrink-0 border border-outline text-on-surface-variant" aria-label="Erledigt: {m.ort}, {m.was}" title="Erledigt" onclick={() => abhaken(m)}>
			<Icon name="check" groesse={22} />
		</button>
	{/if}
</div>
