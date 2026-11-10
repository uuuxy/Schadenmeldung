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
	import LeerZustand from '../components/LeerZustand.svelte';
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
	// Ist die Meldung schon erledigt, ist die Mail überholt, etwa weil der andere Hausmeister schneller
	// war. Dann stehen die Knöpfe erst nach einem bewussten Tipp da, sonst öffnete ein schneller Tipp auf
	// den hervorgehobenen Knopf aus der Mail die Meldung aus Versehen wieder.
	let aendern = $state(false);
	const schonErledigt = $derived(meldung?.status === 'erledigt' && !aendern);

	async function laden() {
		try {
			meldung = await api(`/api/aktion/${token}`);
		} catch (e) {
			fehler = e.message;
		}
	}
	onMount(laden);

	async function setze(status) {
		laeuft = status;
		try {
			// Mit dem gezeigten Stand: Hat inzwischen jemand erledigt, lehnt der Server ab (409), statt die
			// Meldung wieder zu öffnen, und die Seite zeigt den neuen Stand.
			meldung = await api(`/api/aktion/${token}`, { methode: 'POST', json: { status, vorher: meldung.status } });
			getan = status;
			aendern = true; // „Doch anders“ führt dann zurück zu den Knöpfen.
			rueckmeldung();
		} catch (e) {
			if (e.status === 409) {
				aendern = false;
				await laden();
			} else {
				fehler = e.message;
			}
		} finally {
			laeuft = '';
		}
	}
</script>

<KopfLeiste titel="Schadensmeldung" konto={false} />

<main class="mx-auto flex max-w-xl flex-col gap-6 px-4 pt-2 pb-12">
	{#if fehler}
		<LeerZustand icon="schedule" titel="Das hat nicht geklappt" text={fehler}>
			<Button onclick={() => gehe('/', { ersetzen: true })}>App öffnen</Button>
		</LeerZustand>
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
		{:else if schonErledigt}
			<div class="flex flex-col items-center py-4 text-center" in:fade={{ duration: 200 }}>
				<span class="mb-4 grid size-16 place-items-center rounded-xl bg-success-container text-on-success-container"><Icon name="check" groesse={32} /></span>
				<p class="text-title-large text-on-surface">Schon erledigt</p>
				<p class="mt-1 text-body-medium text-on-surface-variant">Hier ist nichts mehr zu tun. Ist doch noch etwas kaputt, lässt sich der Stand ändern.</p>
			</div>
			<Button variant="ghost" onclick={() => (aendern = true)}>Stand ändern</Button>
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
			<p class="text-center text-body-small text-on-surface-variant">„In Arbeit“ zeigt dem Kollegium, dass sich jemand kümmert – ohne Zusage, wann.</p>
		{/if}
		<Button variant="ghost" onclick={() => gehe('/', { ersetzen: true })}><Icon name="open_in_new" groesse={18} />In der App öffnen</Button>
	{/if}
</main>
