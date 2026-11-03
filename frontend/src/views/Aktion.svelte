<script>
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { api } from '../lib/api.js';
	import { gehe } from '../lib/weg.svelte.js';
	import { kategorieNach } from '../lib/kategorien.js';
	import { stati, zwischenstati } from '../lib/status.js';
	import { wann, kurzname } from '../lib/zeit.js';
	import { rueckmeldung } from '../lib/haptik.js';
	import Icon from '../components/Icon.svelte';
	import Ladekreis from '../components/Ladekreis.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Button from '../components/ui/Button.svelte';
	import StatusChip from '../components/ui/StatusChip.svelte';

	/**
	 * Die Seite hinter den Knöpfen der Mail: ohne Anmeldung, der Link selbst berechtigt. Sie führt
	 * nichts beim Öffnen aus, erst beim Antippen: Virenscanner von Mailprogrammen rufen Links
	 * vorab auf und hätten sonst schon „Erledigt“ gedrückt.
	 */
	let { token } = $props();
	const vorschlag = new URLSearchParams(location.search).get('tun');
	// Wie in der App: erst die Zwischenstati, dann „Erledigt“. Die drei Knöpfe der Mail stehen in
	// derselben Reihenfolge, so findet der Hausmeister den angetippten an seinem Platz wieder.
	const angebote = [...zwischenstati, 'erledigt'];

	let meldung = $state(null);
	let fehler = $state('');
	let laeuft = $state('');
	let getan = $state('');

	onMount(async () => {
		try {
			meldung = await api(`/api/aktion/${token}`);
		} catch (e) {
			fehler = e.message;
		}
	});

	async function setze(status) {
		laeuft = status;
		try {
			meldung = await api(`/api/aktion/${token}`, { methode: 'POST', json: { status } });
			getan = status;
			rueckmeldung();
		} catch (e) {
			fehler = e.message;
		} finally {
			laeuft = '';
		}
	}
</script>

<KopfLeiste titel="Schadenmeldung" konto={false} />

<main class="mx-auto flex max-w-xl flex-col gap-6 px-4 pt-2 pb-12">
	{#if fehler}
		<div class="flex flex-col items-center px-4 py-16 text-center" in:fade>
			<span class="mb-5 grid size-20 place-items-center rounded-2xl bg-surface-container-highest text-on-surface-variant"><Icon name="schedule" groesse={40} /></span>
			<p class="text-title-large text-on-surface">Das hat nicht geklappt</p>
			<p class="mt-2 text-body-medium text-on-surface-variant">{fehler}</p>
			<Button class="mt-8" onclick={() => gehe('/', { ersetzen: true })}>App öffnen</Button>
		</div>
	{:else if !meldung}
		<div class="grid place-items-center py-24 text-primary"><Ladekreis groesse={40} /></div>
	{:else}
		<article class="rounded-xl bg-surface-container-low p-4">
			<div class="flex gap-4">
				<span class={['grid size-10 shrink-0 place-items-center rounded-full', kategorieNach(meldung.kategorie).farbe.flaeche]}>
					<Icon name={kategorieNach(meldung.kategorie).icon} groesse={22} class={kategorieNach(meldung.kategorie).farbe.symbol} />
				</span>
				<div class="min-w-0 flex-1">
					<h2 class="text-title-medium text-on-surface">{meldung.ort}</h2>
					{#if meldung.bereich}<p class="text-body-small text-on-surface-variant">{meldung.bereich}</p>{/if}
					<p class="mt-1 text-body-large text-on-surface">{meldung.was}</p>
					{#if meldung.notiz}<p class="mt-0.5 text-body-medium text-on-surface-variant">{meldung.notiz}</p>{/if}
					{#if meldung.foto}<img src={meldung.foto} alt="Foto zur Meldung" class="mt-3 max-h-64 rounded-md object-cover" />{/if}
					<div class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-2">
						<StatusChip status={meldung.status} />
						<span class="text-body-small text-on-surface-variant">{wann(meldung.gemeldetAm)}{meldung.gemeldetVon ? ` · ${kurzname(meldung.gemeldetVon)}` : ''}</span>
					</div>
				</div>
			</div>
		</article>

		{#if getan}
			<div class="flex flex-col items-center py-4 text-center" in:fade={{ duration: 200 }}>
				<span class="mb-4 grid size-16 place-items-center rounded-xl bg-success-container text-on-success-container"><Icon name="check" groesse={32} /></span>
				<p class="text-title-large text-on-surface">„{stati[getan].name}“ ist gesetzt</p>
				<p class="mt-1 text-body-medium text-on-surface-variant">Das Kollegium sieht es sofort.</p>
			</div>
			<Button variant="ghost" onclick={() => (getan = '')}>Doch anders</Button>
		{:else}
			<div class="flex flex-col gap-3">
				{#each angebote as s (s)}
					<Button
						size="lg"
						variant={s === vorschlag ? 'primary' : s === meldung.status ? 'tonal' : 'secondary'}
						disabled={laeuft !== ''}
						onclick={() => setze(s)}
					>
						{#if laeuft === s}<Ladekreis />{:else}<Icon name={s === meldung.status ? 'check' : stati[s].icon} />{/if}
						{stati[s].name}
					</Button>
				{/each}
			</div>
			<p class="text-center text-body-small text-on-surface-variant">„Gesehen“ zeigt dem Kollegium, dass Bescheid gesagt ist – ohne Zusage, wann.</p>
		{/if}
		<Button variant="ghost" onclick={() => gehe('/', { ersetzen: true })}><Icon name="open_in_new" groesse={18} />In der App öffnen</Button>
	{/if}
</main>
