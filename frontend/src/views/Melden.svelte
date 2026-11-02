<script>
	import { tick } from 'svelte';
	import { slide, fade } from 'svelte/transition';
	import { api } from '../lib/api.js';
	import { zustand, uebernimm, darfBearbeiten, darfSehen, legeInAusgang } from '../lib/zustand.svelte.js';
	import { gehe } from '../lib/weg.svelte.js';
	import { zeige } from '../lib/snackbar.svelte.js';
	import { rueckmeldung } from '../lib/haptik.js';
	import { kategorien, kategorieNach } from '../lib/kategorien.js';
	import { verkleinereFoto } from '../lib/foto.js';
	import { wann } from '../lib/zeit.js';
	import { zuletztGewaehlt, merkeRaum } from '../lib/zuletzt.js';
	import Icon from '../components/Icon.svelte';
	import Ladekreis from '../components/Ladekreis.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Button from '../components/ui/Button.svelte';
	import Feld from '../components/ui/Feld.svelte';
	import FilterChips from '../components/ui/FilterChips.svelte';
	import StatusChip from '../components/ui/StatusChip.svelte';
	import Suchpille from '../components/ui/Suchpille.svelte';

	/**
	 * Melden: das Einzige, was das Kollegium sieht. Wo? (QR-Code, zuletzt, Bereich → Raum oder
	 * Suche) → Was? (Art, dann typischer Fall) → Melden. Datum und Name setzt der Server.
	 */
	let { raumCode = '' } = $props();

	const OHNE_BEREICH = 'Weitere Räume';

	let raum = $state(null);
	let freierOrt = $state(false);
	let ortFrei = $state('');
	let suche = $state('');
	let bereich = $state(undefined);
	let kategorie = $state('');
	let fall = $state('');
	let notiz = $state('');
	let notizOffen = $state(false);
	let dringend = $state(false);
	let foto = $state(null);
	let fotoUrl = $state('');
	let sendet = $state(false);
	let fertig = $state(null);
	let zuletzt = $state(zuletztGewaehlt());
	let woAbschnitt = $state();
	let genauer = $state();

	const codeRaum = $derived(raumCode ? zustand.raeume.find((r) => r.code === raumCode) : null);
	$effect(() => {
		if (codeRaum) raum = codeRaum;
	});

	// ── Wo? ──────────────────────────────────────────────────────────────────
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

	// ── Was? ─────────────────────────────────────────────────────────────────
	const art = $derived(kategorie ? kategorieNach(kategorie) : null);
	const freiBeschreiben = $derived(fall === '__anderes' || kategorie === 'sonstiges');
	const ort = $derived(raum ? raum.name : nurFreitext ? ortFrei.trim() : '');
	const was = $derived(freiBeschreiben ? notiz.trim() : fall);
	const bereit = $derived(ort !== '' && kategorie !== '' && was !== '');
	const hierGemeldet = $derived(raum?.id ? zustand.offen.filter((m) => m.raumId === raum.id) : []);
	const hierWartend = $derived(raum?.id ? zustand.ausgang.filter((a) => a.raumId === raum.id) : []);

	// Der Knopf sagt, was noch fehlt, und führt beim Antippen dorthin.
	const knopfText = $derived.by(() => {
		if (bereit) return `Melden · ${ort}`;
		if (!ort) return 'Erst „Wo?“ wählen';
		if (!kategorie) return 'Was ist kaputt? Bitte antippen';
		if (freiBeschreiben) return 'Kurz beschreiben, was kaputt ist';
		return 'Was genau? Bitte antippen';
	});

	// Die Snackbar steht über der Melden-Leiste statt auf ihr.
	$effect(() => {
		if (fertig) return;
		document.body.style.setProperty('--snackbar-abstand', '5.25rem');
		return () => document.body.style.removeProperty('--snackbar-abstand');
	});

	function waehleRaum(r) {
		raum = r;
		suche = '';
		freierOrt = false;
		merkeRaum(r.id);
		zuletzt = zuletztGewaehlt();
		rueckmeldung();
	}

	function ortAusSuche() {
		ortFrei = suche.trim();
		suche = '';
		freierOrt = true;
	}

	async function waehleKategorie(key) {
		if (kategorie === key) return;
		const ersteWahl = !kategorie;
		kategorie = key;
		fall = '';
		notizOffen = false;
		// Beim ersten Mal rollt onintroend nach dem Aufklappen; beim Wechsel ist der Abschnitt schon da.
		if (!ersteWahl) {
			await tick();
			zeigeGenauer();
		}
	}

	// Am Handy liegen die Fälle unter den Kacheln; sie rollen ins Bild, ohne dass jemand wischen muss.
	function zeigeGenauer() {
		genauer?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
	}

	async function fotoGewaehlt(e) {
		const datei = e.currentTarget.files?.[0];
		e.currentTarget.value = '';
		if (!datei) return;
		try {
			foto = await verkleinereFoto(datei);
		} catch {
			foto = datei; // Kann der Browser das Format nicht umrechnen, prüft es der Server.
		}
		if (fotoUrl) URL.revokeObjectURL(fotoUrl);
		fotoUrl = URL.createObjectURL(foto);
	}

	function fotoWeg() {
		if (fotoUrl) URL.revokeObjectURL(fotoUrl);
		foto = null;
		fotoUrl = '';
	}

	function weiter() {
		if (bereit) return melden();
		if (!ort) woAbschnitt?.scrollIntoView({ behavior: 'smooth', block: 'start' });
		else zeigeGenauer();
	}

	async function melden() {
		if (!bereit || sendet) return;
		sendet = true;
		const felder = { kategorie, was };
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
			if (e.status !== 0) {
				zeige(e.message);
				return;
			}
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

	function abschliessen(art) {
		fertig = { art, was, ort, dringend };
		kategorie = '';
		fall = '';
		notiz = '';
		notizOffen = false;
		dringend = false;
		fotoWeg();
		rueckmeldung();
		window.scrollTo({ top: 0 });
	}
</script>

{#snippet ueberschrift(text, erledigt)}
	<h2 class="mb-3 flex items-center gap-2 text-title-medium text-on-surface">
		{text}
		{#if erledigt}<Icon name="check_circle" groesse={20} gefuellt class="text-primary" />{/if}
	</h2>
{/snippet}

{#snippet raumKnopf(r, mitBereich)}
	<button
		class="flex min-h-12 flex-col items-center justify-center rounded-md border border-outline-variant bg-surface-container-lowest px-2 py-2 text-center text-on-surface transition-transform duration-200 active:scale-[0.97]"
		onclick={() => waehleRaum(r)}
	>
		<span class="text-label-large [overflow-wrap:anywhere]">{r.name}</span>
		{#if mitBereich && r.bereich}<span class="mt-0.5 line-clamp-1 text-body-small text-on-surface-variant">{r.bereich}</span>{/if}
	</button>
{/snippet}

<KopfLeiste titel="Schaden melden" ebene="melden" />

{#if fertig}
	<!-- Bestätigung statt Formular: Die Meldung ist sicher angekommen (oder sicher gespeichert). -->
	<main class="mx-auto flex max-w-xl flex-col items-center px-6 pt-16 pb-12 text-center" in:fade={{ duration: 200 }}>
		<span class={['mb-6 grid size-24 place-items-center rounded-2xl', fertig.art === 'gemeldet' ? 'bg-success-container text-on-success-container' : 'bg-surface-container-highest text-on-surface-variant']}>
			<Icon name={fertig.art === 'gemeldet' ? 'task_alt' : 'schedule_send'} groesse={48} />
		</span>
		<h2 class="text-headline-small text-on-surface">{fertig.art === 'gemeldet' ? 'Gemeldet – danke!' : 'Gespeichert'}</h2>
		<p class="mt-2 text-body-large text-on-surface-variant">{fertig.was} · {fertig.ort}</p>
		<p class="mt-4 max-w-sm text-body-medium text-on-surface-variant">
			{#if fertig.art === 'ausgang'}
				Gerade ist kein Netz. Die Meldung geht automatisch raus, sobald wieder Verbindung besteht.
			{:else if fertig.dringend}
				Der Hausmeister bekommt sie sofort per Mail.
			{:else}
				Sie steht jetzt auf der Aufgabenliste.
			{/if}
		</p>
		<div class="mt-10 flex flex-col gap-3 self-stretch">
			<Button size="lg" onclick={() => (fertig = null)}><Icon name="add" />Noch etwas melden</Button>
			{#if darfSehen()}
				<Button variant="ghost" onclick={() => gehe('/')}>Zur Übersicht</Button>
			{/if}
		</div>
	</main>
{:else}
	<main class="mx-auto flex max-w-3xl flex-col gap-8 px-4 pt-2 pb-36">
		<!-- Wo? -->
		<section bind:this={woAbschnitt} class="scroll-mt-20">
			{@render ueberschrift('Wo?', ort !== '')}
			{#if raumCode && !codeRaum && zustand.raeume.length && !raum}
				<p class="mb-3 rounded-md bg-warning-container px-4 py-3 text-body-medium text-on-warning-container">
					Dieser QR-Code gehört zu keinem Raum mehr. Bitte den Raum auswählen.
				</p>
			{/if}

			{#if raum}
				<button class="flex w-full items-center gap-4 rounded-xl bg-secondary-container p-4 text-left text-on-secondary-container" onclick={() => (raum = null)}>
					<Icon name="location_on" gefuellt groesse={28} />
					<span class="min-w-0 flex-1">
						<span class="block truncate text-title-large">{raum.name}</span>
						{#if raum.bereich}<span class="block truncate text-body-medium">{raum.bereich}</span>{/if}
					</span>
					<span class="shrink-0 text-label-large">Ändern</span>
				</button>
			{:else if nurFreitext}
				<Feld etikett="Wo ist der Schaden?" placeholder="z. B. Schulhof, Treppenhaus Nord" bind:wert={ortFrei} autocomplete="off" enterkeyhint="next" maxlength="80" />
				{#if zustand.raeume.length}
					<Button variant="ghost" class="mt-2 -ml-4" onclick={() => (freierOrt = false)}><Icon name="meeting_room" groesse={18} />Doch einen Raum wählen</Button>
				{:else if darfBearbeiten()}
					<p class="mt-2 text-body-small text-on-surface-variant">Tipp: Unter „Räume & Personen“ (Kontomenü) die Raumliste anlegen – dann genügt hier ein Tipp.</p>
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

			{#if hierGemeldet.length || hierWartend.length}
				<div class="mt-4 overflow-hidden rounded-xl bg-surface-container-low" transition:slide={{ duration: 250 }}>
					<p class="px-4 pt-3 pb-1 text-label-large text-on-surface-variant">Hier schon gemeldet</p>
					<ul>
						{#each hierWartend as a (a.id)}
							<li class="flex items-center gap-3 px-4 py-3">
								<Icon name={kategorieNach(a.kategorie).icon} groesse={20} class={kategorieNach(a.kategorie).farbe.symbol} />
								<span class="min-w-0 flex-1 truncate text-body-large text-on-surface">{a.was}</span>
								<span class="inline-flex h-6 items-center gap-1 rounded-sm bg-surface-container px-2 text-label-medium text-on-surface-variant"><Icon name="schedule_send" groesse={16} />Wartet auf Netz</span>
							</li>
						{/each}
						{#each hierGemeldet as m (m.id)}
							<li class="flex items-center gap-3 px-4 py-3">
								<Icon name={kategorieNach(m.kategorie).icon} groesse={20} class={kategorieNach(m.kategorie).farbe.symbol} />
								<span class="min-w-0 flex-1">
									<span class="block truncate text-body-large text-on-surface">{m.was}</span>
									<span class="block text-body-small text-on-surface-variant">{wann(m.gemeldetAm)}</span>
								</span>
								<StatusChip status={m.status} />
							</li>
						{/each}
					</ul>
				</div>
			{/if}
		</section>

		<!-- Was ist kaputt? Jede Art leuchtet ausgewählt in ihrer eigenen Farbe (kat-* in rollen.css),
		     wie Geräte-Kacheln in Google Home; das Symbol trägt die Farbe schon vorher. -->
		<section>
			{@render ueberschrift('Was ist kaputt?', was !== '')}
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-4" role="group" aria-label="Art des Schadens">
				{#each kategorien as k (k.key)}
					{@const aktiv = kategorie === k.key}
					<button
						class={[
							'relative flex min-h-28 flex-col justify-between gap-3 rounded-xl p-4 text-left transition-colors duration-200 active:scale-[0.98]',
							aktiv ? k.farbe.flaeche : 'bg-surface-container-low text-on-surface'
						]}
						aria-pressed={aktiv}
						onclick={() => waehleKategorie(k.key)}
					>
						<Icon name={k.icon} gefuellt={aktiv} groesse={28} class={k.farbe.symbol} />
						<span class="text-title-medium leading-tight">{k.name}</span>
						{#if aktiv}<Icon name="check_circle" gefuellt groesse={22} class={['absolute top-3 right-3', k.farbe.symbol]} />{/if}
					</button>
				{/each}
			</div>
		</section>

		{#if art}
			<section bind:this={genauer} class="scroll-mb-48" transition:slide={{ duration: 300 }} onintroend={zeigeGenauer}>
				{#if art.faelle.length}
					{@render ueberschrift('Genauer?', was !== '')}
					<FilterChips
						etikett="Was genau"
						optionen={[...art.faelle.map((f) => ({ wert: f, text: f })), { wert: '__anderes', text: 'Anderes …' }]}
						wert={fall || null}
						onwahl={(f) => (fall = f ?? '')}
					/>
				{:else}
					{@render ueberschrift('Kurz beschreiben', was !== '')}
				{/if}

				{#if freiBeschreiben}
					<div class="mt-4" transition:slide={{ duration: 200 }}>
						<Feld etikett="Was ist kaputt?" bind:wert={notiz} maxlength="160" enterkeyhint="done" autocomplete="off" />
					</div>
				{/if}

				<!-- Freiwillig: Foto, Notiz, Dringend. Assist-Chips (M3) mit Symbol in primary. -->
				<div class="mt-6 flex flex-wrap items-center gap-x-2 gap-y-4">
					<label class="schicht treffer flex h-8 cursor-pointer items-center gap-2 rounded-sm border border-outline pr-4 pl-2 text-label-large text-on-surface">
						<Icon name="photo_camera" groesse={18} class="text-primary" />{foto ? 'Anderes Foto' : 'Foto'}
						<input type="file" accept="image/*" capture="environment" class="sr-only" onchange={fotoGewaehlt} />
					</label>
					{#if !freiBeschreiben}
						<button class={['treffer flex h-8 items-center gap-2 rounded-sm border pr-4 pl-2 text-label-large transition-colors', notizOffen ? 'border-transparent bg-secondary-container text-on-secondary-container' : 'border-outline text-on-surface']} aria-pressed={notizOffen} onclick={() => (notizOffen = !notizOffen)}>
							<Icon name={notizOffen ? 'check' : 'edit_note'} groesse={18} class={notizOffen ? '' : 'text-primary'} />Notiz
						</button>
					{/if}
					<button class={['treffer flex h-8 items-center gap-2 rounded-sm border pr-4 pl-2 text-label-large transition-colors', dringend ? 'border-transparent bg-error-container text-on-error-container' : 'border-outline text-on-surface']} aria-pressed={dringend} onclick={() => (dringend = !dringend)}>
						<Icon name="warning" groesse={18} gefuellt={dringend} class={dringend ? '' : 'text-error'} />Dringend / Gefahr
					</button>
				</div>

				{#if fotoUrl}
					<div class="relative mt-4 inline-block" transition:slide={{ duration: 200 }}>
						<img src={fotoUrl} alt="Gewähltes Foto" class="h-28 w-28 rounded-md object-cover" />
						<button class="icon-btn absolute -top-2 -right-2 size-8 bg-inverse-surface text-inverse-on-surface" aria-label="Foto entfernen" onclick={fotoWeg}>
							<Icon name="close" groesse={18} />
						</button>
					</div>
				{/if}

				{#if notizOffen && !freiBeschreiben}
					<div class="mt-4" transition:slide={{ duration: 200 }}>
						<Feld etikett="Notiz (freiwillig)" bind:wert={notiz} mehrzeilig zeilen={2} maxlength="600" hinweis="Bitte keine Namen von Schülerinnen und Schülern." />
					</div>
				{/if}
			</section>
		{/if}
	</main>

	<!-- Die Hauptaktion unten in Daumenreichweite, auf einer Leiste in surface-container (M3 Bottom
	     App Bar) statt mit Schatten über dem Inhalt. -->
	<div class="fixed inset-x-0 bottom-0 z-20 bg-surface-container px-4 pt-3 pb-[calc(env(safe-area-inset-bottom)+0.75rem)]">
		<Button size="lg" variant={bereit ? 'primary' : 'secondary'} class="mx-auto w-full max-w-3xl" disabled={sendet} onclick={weiter}>
			{#if sendet}
				<Ladekreis />
			{:else}
				{#if bereit}<Icon name="send" />{/if}
				<span class="truncate">{knopfText}</span>
			{/if}
		</Button>
	</div>
{/if}
