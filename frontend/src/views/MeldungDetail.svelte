<script>
	import { slide } from 'svelte/transition';
	import { zustand, darfBearbeiten } from '../lib/zustand.svelte.js';
	import { abhaken, loeschen, vermerkSpeichern, wiederOeffnen } from '../lib/meldung.js';
	import { kategorieNach } from '../lib/kategorien.js';
	import { wann, kurzname } from '../lib/zeit.js';
	import { zurueck } from '../lib/weg.svelte.js';
	import Icon from '../components/Icon.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Button from '../components/ui/Button.svelte';
	import Feld from '../components/ui/Feld.svelte';
	import StatusChip from '../components/ui/StatusChip.svelte';
	import Stand from '../components/meldung/Stand.svelte';

	/**
	 * Eine Meldung mit allem, was dazugehört: Foto in voller Breite, Stand, interner Vermerk. Die
	 * Hauptaktion „Erledigt“ liegt unten in Daumenreichweite wie „Melden“ im Formular. Die Seite
	 * liest die Meldung aus der Liste und bleibt damit live.
	 */
	let { id } = $props();
	const m = $derived(zustand.offen.find((x) => x.id === id) ?? zustand.erledigt.find((x) => x.id === id));
	const art = $derived(m ? kategorieNach(m.kategorie) : null);
	const erledigt = $derived(m?.status === 'erledigt');
	let vermerk = $state(null);
	const vermerkGeaendert = $derived(m && vermerk !== null && vermerk.trim() !== m.antwort);

	$effect(() => {
		if (m && vermerk === null) vermerk = m.antwort;
	});
	$effect(() => {
		document.body.style.setProperty('--snackbar-abstand', darfBearbeiten() ? '5.25rem' : '0px');
		return () => document.body.style.removeProperty('--snackbar-abstand');
	});

	async function fertig() {
		const meldung = m;
		zurueck('/');
		await abhaken(meldung);
	}

	async function weg() {
		if (await loeschen(m)) zurueck('/');
	}
</script>

<KopfLeiste titel={m?.ort ?? 'Meldung'} zurueck={() => zurueck('/')} />

{#if !m}
	<main class="mx-auto flex max-w-xl flex-col items-center px-6 py-16 text-center">
		<span class="mb-5 grid size-20 place-items-center rounded-2xl bg-surface-container-highest text-on-surface-variant"><Icon name="task_alt" groesse={40} /></span>
		<p class="text-title-large text-on-surface">Diese Meldung gibt es hier nicht mehr</p>
		<p class="mt-2 text-body-medium text-on-surface-variant">Sie wurde gelöscht oder ist seit mehr als 30 Tagen erledigt.</p>
		<Button class="mt-8" onclick={() => zurueck('/')}>Zur Übersicht</Button>
	</main>
{:else}
	<main class="mx-auto flex max-w-3xl flex-col gap-6 px-4 pt-2 pb-32">
		<div class="flex items-start gap-4">
			<span class={['grid size-12 shrink-0 place-items-center rounded-full', art.farbe.flaeche]}><Icon name={art.icon} groesse={24} class={art.farbe.symbol} /></span>
			<div class="min-w-0 flex-1">
				<p class="text-body-medium text-on-surface-variant">{art.name}{m.bereich ? ` · ${m.bereich}` : ''}</p>
				<h2 class="mt-0.5 text-headline-small text-on-surface">{m.was}</h2>
				{#if m.notiz}<p class="mt-2 text-body-large text-on-surface-variant">{m.notiz}</p>{/if}
				<div class="mt-3 flex flex-wrap items-center gap-2">
					<StatusChip status={m.status} />
					{#if m.dringend && !erledigt}
						<span class="inline-flex h-6 items-center gap-1 rounded-sm bg-error-container px-2 text-label-medium text-on-error-container"><Icon name="warning" groesse={16} gefuellt />Dringend</span>
					{/if}
				</div>
			</div>
		</div>

		{#if m.foto}
			<a href={m.foto} target="_blank" rel="noopener" class="block overflow-hidden rounded-xl bg-surface-container-highest" aria-label="Foto in voller Größe öffnen">
				<img src={m.foto} alt="Foto zur Meldung in {m.ort}" class="max-h-[60dvh] w-full object-contain" />
			</a>
		{/if}

		<dl class="flex flex-col gap-0.5 text-body-medium">
			{#each [['Gemeldet', `${wann(m.gemeldetAm)}${m.gemeldetVon ? ` · ${kurzname(m.gemeldetVon)}` : ''}`], ...(erledigt && m.erledigtAm ? [['Erledigt', wann(m.erledigtAm)]] : [])] as [name, wert] (name)}
				<div class="flex justify-between gap-4 rounded-xs bg-surface-container-low px-4 py-3 first:rounded-t-lg last:rounded-b-lg">
					<dt class="text-on-surface-variant">{name}</dt>
					<dd class="text-right text-on-surface">{wert}</dd>
				</div>
			{/each}
		</dl>

		{#if darfBearbeiten() && !erledigt}
			<section>
				<h3 class="mb-3 text-title-small text-on-surface-variant">Stand</h3>
				<Stand meldung={m} />
				<p class="mt-2 text-body-small text-on-surface-variant">„Gesehen“ zeigt dem Kollegium, dass Bescheid gesagt ist – ohne Zusage, wann. Ein zweiter Tipp nimmt den Stand zurück.</p>
			</section>
		{/if}

		{#if darfBearbeiten()}
			<section>
				<Feld etikett="Interner Vermerk" bind:wert={vermerk} mehrzeilig zeilen={2} maxlength="300" hinweis="Sehen nur Hausmeister, Leitung und Verwaltung, nicht das Kollegium." />
				{#if vermerkGeaendert}
					<div class="mt-3 flex justify-end gap-2" transition:slide={{ duration: 200 }}>
						<Button variant="ghost" onclick={() => (vermerk = m.antwort)}>Verwerfen</Button>
						<Button variant="tonal" onclick={() => vermerkSpeichern(m, vermerk)}>Vermerk speichern</Button>
					</div>
				{/if}
			</section>
			<Button variant="warnend" class="-ml-3 self-start" onclick={weg}><Icon name="delete" groesse={18} />Meldung löschen</Button>
		{:else if m.antwort}
			<section>
				<h3 class="mb-2 text-title-small text-on-surface-variant">Vermerk</h3>
				<p class="rounded-lg bg-surface-container-low px-4 py-3 text-body-large text-on-surface">{m.antwort}</p>
			</section>
		{/if}
	</main>

	{#if darfBearbeiten()}
		<div class="fixed inset-x-0 bottom-0 z-20 bg-surface-container px-4 pt-3 pb-[calc(env(safe-area-inset-bottom)+0.75rem)]">
			{#if erledigt}
				<Button size="lg" variant="tonal" class="mx-auto w-full max-w-3xl" onclick={() => wiederOeffnen(m)}><Icon name="undo" />Wieder öffnen</Button>
			{:else}
				<Button size="lg" class="mx-auto w-full max-w-3xl" onclick={fertig}><Icon name="check" />Erledigt</Button>
			{/if}
		</div>
	{/if}
{/if}
