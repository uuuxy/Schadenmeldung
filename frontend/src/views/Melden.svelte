<script>
	import { tick } from 'svelte';
	import { slide } from 'svelte/transition';
	import { api } from '../lib/api.js';
	import { zustand, uebernimm, legeInAusgang } from '../lib/zustand.svelte.js';
	import { neueKennung } from '../lib/ausgang.js';
	import { zeige } from '../lib/snackbar.svelte.js';
	import { rueckmeldung } from '../lib/haptik.js';
	import { kategorieNach, sichtbareArten } from '../lib/kategorien.js';
	import Icon from '../components/Icon.svelte';
	import Ladekreis from '../components/Ladekreis.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Button from '../components/ui/Button.svelte';
	import Frage from '../components/melden/Frage.svelte';
	import WoWahl from '../components/melden/WoWahl.svelte';
	import ArtWahl from '../components/melden/ArtWahl.svelte';
	import FallWahl from '../components/melden/FallWahl.svelte';
	import Gemeldet from '../components/melden/Gemeldet.svelte';
	import AlsApp from '../components/melden/AlsApp.svelte';

	/**
	 * Melden: das Einzige, was das Kollegium sieht. Drei Fragen untereinander – Wo? Was ist kaputt?
	 * Was genau? Jede beantwortete schrumpft auf eine Zeile mit „Ändern“, so steht die offene Frage
	 * oben im Bild. Datum und Name setzt der Server.
	 */
	let { raumCode = '' } = $props();

	let raum = $state(null);
	let freierOrt = $state(false);
	let ortFrei = $state('');
	let kategorie = $state('');
	let fall = $state('');
	let notiz = $state('');
	let dringend = $state(false);
	let foto = $state(null);
	let sendet = $state(false);
	let fertig = $state(null);
	// Welche Frage noch fehlt, nachdem jemand zu früh auf „Melden“ getippt hat.
	let fehlt = $state('');
	const abschnitt = $state({ wo: null, was: null, genau: null });

	const codeRaum = $derived(raumCode ? zustand.raeume.find((r) => r.code === raumCode) : null);
	$effect(() => {
		if (codeRaum) raum = codeRaum;
	});

	const arten = $derived(sichtbareArten(zustand.einstellungen.artenAus));
	const art = $derived(kategorie ? kategorieNach(kategorie) : null);
	const freiBeschreiben = $derived(fall === '__anderes' || kategorie === 'sonstiges');
	const ort = $derived(raum ? raum.name : zustand.raeume.length === 0 || freierOrt ? ortFrei.trim() : '');
	const was = $derived(freiBeschreiben ? notiz.trim() : fall);
	const offen = $derived(!ort ? 'wo' : !kategorie ? 'was' : !was ? 'genau' : '');
	const hinweise = $derived({
		wo: 'Bitte einen Raum antippen, suchen oder einen Ort eingeben.',
		was: 'Bitte eine Kachel antippen.',
		genau: freiBeschreiben ? 'Bitte kurz beschreiben, was kaputt ist.' : 'Bitte antippen, was genau kaputt ist.'
	});

	// Ein Hinweis gilt nur, bis die Frage beantwortet ist.
	$effect(() => {
		if (fehlt && fehlt !== offen) fehlt = '';
	});

	// Die Snackbar steht über der Leiste mit dem Knopf statt auf ihr.
	$effect(() => {
		if (fertig) return;
		document.body.style.setProperty('--snackbar-abstand', '5.25rem');
		return () => document.body.style.removeProperty('--snackbar-abstand');
	});

	// Nach jeder Antwort rollt die beantwortete Frage nach oben; darunter steht die nächste.
	async function weiterNach(name) {
		await tick();
		abschnitt[name]?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	function waehleKategorie(key) {
		kategorie = key;
		fall = '';
		notiz = '';
		if (key) rueckmeldung();
		weiterNach('was');
	}

	function knopf() {
		if (!offen) return melden();
		fehlt = offen;
		abschnitt[offen]?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	async function melden() {
		if (sendet) return;
		sendet = true;
		const felder = { kategorie, was, kennung: neueKennung() };
		if (raum?.id) felder.raumId = String(raum.id);
		else felder.ort = ort;
		if (!freiBeschreiben && notiz.trim()) felder.notiz = notiz.trim();
		if (dringend) felder.dringend = '1';
		const formular = new FormData();
		for (const [schluessel, wert] of Object.entries(felder)) formular.set(schluessel, wert);
		if (foto) formular.set('foto', foto, 'foto.jpg');
		try {
			uebernimm(await api('/api/meldungen', { methode: 'POST', formular }));
			abschliessen('gemeldet');
		} catch (e) {
			if (e.status !== 0) return zeige(e.message);
			// Funkloch: nicht verwerfen, sondern im Ausgang ablegen und später selbst nachreichen.
			try {
				await legeInAusgang(felder, foto, { ort, raumId: raum?.id ?? null, kategorie, was });
				zustand.offline = true;
				abschliessen('ausgang');
			} catch {
				zeige(e.message);
			}
		} finally {
			sendet = false;
		}
	}

	function abschliessen(artDerMeldung) {
		fertig = { art: artDerMeldung, was, ort, dringend };
		kategorie = '';
		fall = '';
		notiz = '';
		dringend = false;
		foto = null;
		rueckmeldung();
		window.scrollTo({ top: 0 });
	}
</script>

<KopfLeiste titel="Schaden melden" ebene="melden" />

{#if fertig}
	<Gemeldet {fertig} onweiter={() => (fertig = null)} />
{:else}
	<main class="mx-auto flex max-w-3xl flex-col gap-8 px-4 pt-2 pb-36">
		<AlsApp />
		<Frage titel="Wo?" erledigt={ort !== ''} hinweis={fehlt === 'wo' ? hinweise.wo : ''} bind:element={abschnitt.wo}>
			<WoWahl bind:raum bind:ortFrei bind:freierOrt codeUnbekannt={!!raumCode && !codeRaum && zustand.raeume.length > 0} ongewaehlt={() => weiterNach('wo')} />
		</Frage>

		<Frage titel="Was ist kaputt?" erledigt={kategorie !== ''} hinweis={fehlt === 'was' ? hinweise.was : ''} bind:element={abschnitt.was}>
			<ArtWahl {arten} {kategorie} onwahl={waehleKategorie} />
		</Frage>

		{#if art}
			<div transition:slide={{ duration: 250 }}>
				<Frage titel={art.faelle.length ? 'Was genau?' : 'Kurz beschreiben'} erledigt={was !== ''} hinweis={fehlt === 'genau' ? hinweise.genau : ''} bind:element={abschnitt.genau}>
					<FallWahl {art} {freiBeschreiben} bind:fall bind:notiz bind:dringend bind:foto />
				</Frage>
			</div>
		{/if}
	</main>

	<!-- Die Hauptaktion unten in Daumenreichweite, auf einer angedockten Leiste (M3 docked toolbar mit
	     eingebetteter Hauptaktion). Solange etwas fehlt, sieht der Knopf gesperrt aus, bleibt aber
	     antippbar und zeigt dann, welche Frage offen ist (aria-disabled statt disabled). -->
	<div class="fixed inset-x-0 bottom-0 z-20 bg-surface-container px-4 pt-3 pb-[calc(env(safe-area-inset-bottom)+0.75rem)]">
		<Button size="lg" class="mx-auto w-full max-w-3xl" aria-disabled={offen !== ''} disabled={sendet} onclick={knopf}>
			{#if sendet}
				<Ladekreis />
			{:else}
				<Icon name="send" />
				<span class="truncate">{ort ? `Melden · ${ort}` : 'Melden'}</span>
			{/if}
		</Button>
	</div>
{/if}
