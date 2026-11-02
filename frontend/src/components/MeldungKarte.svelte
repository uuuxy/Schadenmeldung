<script>
	import Icon from './Icon.svelte';
	import Blatt from './Blatt.svelte';
	import Button from './ui/Button.svelte';
	import Feld from './ui/Feld.svelte';
	import StatusChip from './ui/StatusChip.svelte';
	import { api } from '../lib/api.js';
	import { zustand, darfBearbeiten, istAdmin, setzeStatus, uebernimm } from '../lib/zustand.svelte.js';
	import { kategorieNach } from '../lib/kategorien.js';
	import { stati, zwischenstati, schnellantworten } from '../lib/status.js';
	import { wann, kurzname } from '../lib/zeit.js';
	import { zeige } from '../lib/snackbar.svelte.js';
	import { rueckmeldung } from '../lib/haptik.js';

	/** Eine Meldung als M3-Karte. Hausmeister und Verwaltung bearbeiten sie, die Leitung liest nur. */
	let { meldung } = $props();

	const art = $derived(kategorieNach(meldung.kategorie));
	const erledigt = $derived(meldung.status === 'erledigt');
	let fotoOffen = $state(false);
	let antwortOffen = $state(false);
	let antwort = $state('');
	let laeuft = $state(false);

	async function setze(aenderung, text) {
		laeuft = true;
		try {
			await setzeStatus(meldung.id, aenderung);
			if (text) zeige(text);
		} catch (e) {
			zeige(e.message);
		} finally {
			laeuft = false;
		}
	}

	// Ein zweiter Tipp auf den gewählten Zwischenstand nimmt ihn zurück.
	function zwischen(status) {
		setze({ status: meldung.status === status ? 'offen' : status });
	}

	// Sofort abhaken, ohne auf den Server zu warten; schlägt es fehl, kommt die Karte zurück.
	function abhaken() {
		const vorher = $state.snapshot(meldung);
		uebernimm({ ...vorher, status: 'erledigt', erledigtAm: new Date().toISOString() });
		rueckmeldung();
		setzeStatus(vorher.id, { status: 'erledigt' })
			.then(() =>
				zeige(`Erledigt: ${vorher.ort}`, {
					name: 'Rückgängig',
					tun: () => setzeStatus(vorher.id, { status: vorher.status }).catch((e) => zeige(e.message))
				})
			)
			.catch((e) => {
				uebernimm(vorher);
				zeige(e.message);
			});
	}

	function antwortSchreiben() {
		antwort = meldung.antwort;
		antwortOffen = true;
	}

	async function antwortSpeichern() {
		await setze({ antwort }, 'Antwort gespeichert');
		antwortOffen = false;
	}

	async function loeschen() {
		if (!confirm(`„${meldung.was}“ (${meldung.ort}) endgültig löschen?`)) return;
		try {
			await api(`/api/meldungen/${meldung.id}`, { methode: 'DELETE' });
			zustand.offen = zustand.offen.filter((m) => m.id !== meldung.id);
			zustand.erledigt = zustand.erledigt.filter((m) => m.id !== meldung.id);
			zeige('Meldung gelöscht');
		} catch (e) {
			zeige(e.message);
		}
	}
</script>

<article class={['rounded-xl bg-surface-container-low p-4', meldung.dringend && !erledigt && 'border-l-4 border-error']}>
	<div class="flex gap-4">
		<!-- Dieselbe Farbe der Art wie im Formular: Beim Rundgang erkennt man Licht, Wasser … am Blick. -->
		<span class={['grid size-10 shrink-0 place-items-center rounded-full', erledigt ? 'bg-success-container text-on-success-container' : art.farbe.flaeche]}>
			<Icon name={erledigt ? 'check' : art.icon} groesse={22} class={erledigt ? '' : art.farbe.symbol} />
		</span>

		<div class="min-w-0 flex-1">
			<div class="flex items-start gap-2">
				<div class="min-w-0 flex-1">
					<h3 class="truncate text-title-medium text-on-surface">{meldung.ort}</h3>
					{#if meldung.bereich}<p class="truncate text-body-small text-on-surface-variant">{meldung.bereich}</p>{/if}
				</div>
				{#if meldung.dringend && !erledigt}
					<span class="inline-flex h-6 shrink-0 items-center gap-1 rounded-sm bg-error-container px-2 text-label-medium text-on-error-container">
						<Icon name="warning" groesse={16} gefuellt />Dringend
					</span>
				{/if}
			</div>

			<p class="mt-1 text-body-large text-on-surface">{meldung.was}</p>
			{#if meldung.notiz}<p class="mt-0.5 text-body-medium text-on-surface-variant">{meldung.notiz}</p>{/if}

			{#if meldung.foto}
				<button class="ohne-schicht mt-3 block overflow-hidden rounded-md" onclick={() => (fotoOffen = true)} aria-label="Foto vergrößern">
					<img src={meldung.foto} alt="" loading="lazy" decoding="async" class="h-32 w-44 object-cover" />
				</button>
			{/if}

			<div class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-2">
				<StatusChip status={meldung.status} />
				<span class="text-body-small text-on-surface-variant">
					{erledigt && meldung.erledigtAm ? `erledigt ${wann(meldung.erledigtAm)}` : wann(meldung.gemeldetAm)}{meldung.gemeldetVon ? ` · ${kurzname(meldung.gemeldetVon)}` : ''}
				</span>
			</div>

			{#if meldung.antwort}
				<p class="mt-3 rounded-md bg-surface-container-high px-3 py-2 text-body-medium text-on-surface">
					<span class="text-on-surface-variant">Hausmeister:</span>
					{meldung.antwort}
				</p>
			{/if}
		</div>
	</div>

	{#if darfBearbeiten()}
		<div class="mt-4 border-t border-outline-variant pt-3">
			{#if erledigt}
				<div class="flex items-center justify-end gap-1">
					{#if istAdmin()}<Button variant="symbol" aria-label="Meldung löschen" onclick={loeschen}><Icon name="delete" /></Button>{/if}
					<Button variant="ghost" disabled={laeuft} onclick={() => setze({ status: 'offen' }, 'Wieder offen')}><Icon name="undo" groesse={18} />Wieder öffnen</Button>
				</div>
			{:else}
				<!-- M3 Expressive „connected button group“: drei Zwischenstände in einer Reihe, der gewählte
				     wird zur Pille und trägt ein Häkchen. -->
				<div class="grid grid-cols-3 gap-0.5" role="group" aria-label="Zwischenstand">
					{#each zwischenstati as s, i (s)}
						{@const gewaehlt = meldung.status === s}
						<button
							class={[
								'treffer flex h-10 min-w-0 items-center justify-center gap-1.5 px-2 text-label-large transition-[background-color,color,border-radius] duration-300',
								gewaehlt
									? 'rounded-full bg-secondary-container text-on-secondary-container'
									: ['bg-surface-container-highest text-on-surface-variant', i === 0 ? 'rounded-l-full rounded-r-sm' : i === 2 ? 'rounded-l-sm rounded-r-full' : 'rounded-sm']
							]}
							aria-pressed={gewaehlt}
							disabled={laeuft}
							onclick={() => zwischen(s)}
						>
							<Icon name={gewaehlt ? 'check' : stati[s].icon} groesse={18} /><span class="truncate">{stati[s].knapp}</span>
						</button>
					{/each}
				</div>
				<div class="mt-3 flex items-center gap-1">
					<Button variant="ghost" class="-ml-2" onclick={antwortSchreiben}><Icon name="add_comment" groesse={18} />Antwort</Button>
					{#if istAdmin()}<Button variant="symbol" aria-label="Meldung löschen" onclick={loeschen}><Icon name="delete" /></Button>{/if}
					<Button variant="success" class="ml-auto pr-5 pl-4" onclick={abhaken}><Icon name="check" groesse={20} />Erledigt</Button>
				</div>
			{/if}
		</div>
	{/if}
</article>

{#if meldung.foto}
	<Blatt bind:offen={fotoOffen} label="Foto zur Meldung">
		<img src={meldung.foto} alt="Foto zur Meldung in {meldung.ort}" class="max-h-[72dvh] w-full bg-surface-container-highest object-contain" />
		<div class="flex justify-end p-3"><Button variant="ghost" onclick={() => (fotoOffen = false)}>Schließen</Button></div>
	</Blatt>
{/if}

{#if darfBearbeiten()}
	<Blatt bind:offen={antwortOffen} titel="Antwort ans Kollegium">
		<div class="flex flex-col gap-4 px-6 pb-6">
			<div class="flex flex-wrap gap-x-2 gap-y-4">
				{#each schnellantworten as text (text)}
					<button class="treffer h-8 rounded-sm border border-outline px-3 text-label-large text-on-surface-variant" onclick={() => (antwort = text)}>{text}</button>
				{/each}
			</div>
			<Feld etikett="Antwort" bind:wert={antwort} mehrzeilig maxlength="300" />
			<div class="flex justify-end gap-2">
				<Button variant="ghost" onclick={() => (antwortOffen = false)}>Abbrechen</Button>
				<Button disabled={laeuft} onclick={antwortSpeichern}>Speichern</Button>
			</div>
		</div>
	</Blatt>
{/if}
