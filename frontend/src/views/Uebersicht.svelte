<script>
	import { flip } from 'svelte/animate';
	import { fade, slide } from 'svelte/transition';
	import { zustand, darfBearbeiten } from '../lib/zustand.svelte.js';
	import { auffrischen } from '../lib/nachreichen.js';
	import Icon from '../components/Icon.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Ziehen from '../components/Ziehen.svelte';
	import Reiter from '../components/ui/Reiter.svelte';
	import Button from '../components/ui/Button.svelte';
	import MenueChip from '../components/ui/MenueChip.svelte';
	import MeldungZeile from '../components/uebersicht/MeldungZeile.svelte';
	import Export from '../components/uebersicht/Export.svelte';
	import MailStoerung from '../components/uebersicht/MailStoerung.svelte';
	import { kategorien } from '../lib/kategorien.js';

	/**
	 * Die Übersicht für Hausmeister, Leitung, Verwaltung und Admins: eine schlanke Liste, Dringendes
	 * unter eigener Überschrift oben, alles andere neueste zuerst. Ein Tipp öffnet die Meldung mit allem
	 * Weiteren (M3 list-detail: am Handy immer nur eine der beiden Ansichten). Filtern lässt sie sich
	 * nach Bereich und Art, je ein Chip mit Blatt statt Reihen von Chips (M3). Reiter und Filter merkt
	 * sich die Sitzung, damit „Zurück“ dort landet.
	 */
	const merker = (() => {
		try {
			return JSON.parse(sessionStorage.getItem('schaden.uebersicht') ?? '{}');
		} catch {
			return {};
		}
	})();
	let reiter = $state(merker.reiter ?? 'offen');
	let bereich = $state(merker.bereich ?? null);
	let art = $state(merker.art ?? null);

	const quelle = $derived(reiter === 'offen' ? zustand.offen : zustand.erledigt);
	// Bereiche in der Reihenfolge der Raumliste (EG vor 1. OG), nur die mit Meldungen.
	const bereiche = $derived.by(() => {
		const mitMeldung = new Set(quelle.map((m) => m.bereich).filter(Boolean));
		return [...new Set(zustand.raeume.map((r) => r.bereich))].filter((b) => mitMeldung.delete(b)).concat([...mitMeldung]);
	});
	// Arten in der Reihenfolge der Kacheln, nur die mit Meldungen; auf dem Chip das erste Wort.
	const arten = $derived(
		kategorien
			.filter((k) => quelle.some((m) => m.kategorie === k.key))
			.map((k) => ({ wert: k.key, text: k.name, kurz: k.name.split(' & ')[0], icon: k.icon, symbol: k.farbe.symbol, flaeche: k.farbe.flaeche }))
	);
	const liste = $derived(quelle.filter((m) => (!bereich || m.bereich === bereich) && (!art || m.kategorie === art)));
	const zeigeBereich = $derived(bereiche.length > 1 || !!bereich);
	const zeigeArt = $derived(arten.length > 1 || !!art);

	// Dringendes steht unter eigener Überschrift oben, damit eine Gefahr nicht unter neueren Meldungen
	// verschwindet, und die Überschrift erklärt, warum. Eine Wahl der Sortierung gibt es bewusst nicht:
	// eine gute Reihenfolge statt einer Einstellung, wie bei Gmail.
	const nachZeit = (a, b) => new Date(b.gemeldetAm) - new Date(a.gemeldetAm);
	const abschnitte = $derived.by(() => {
		if (reiter !== 'offen') return [{ titel: '', meldungen: liste }];
		const dringend = liste.filter((m) => m.dringend).sort(nachZeit);
		const weitere = liste.filter((m) => !m.dringend).sort(nachZeit);
		if (!dringend.length) return [{ titel: '', meldungen: weitere }];
		return [{ titel: 'Dringend', meldungen: dringend }, ...(weitere.length ? [{ titel: 'Weitere, neueste zuerst', meldungen: weitere }] : [])];
	});

	$effect(() => {
		if (bereich && !bereiche.includes(bereich)) bereich = null;
	});
	$effect(() => {
		if (art && !arten.some((a) => a.wert === art)) art = null;
	});
	$effect(() => {
		try {
			sessionStorage.setItem('schaden.uebersicht', JSON.stringify({ reiter, bereich, art }));
		} catch {
			/* Ohne Speicher beginnt die Übersicht eben bei „Offen“. */
		}
	});
</script>

<KopfLeiste titel={darfBearbeiten() ? 'Aufgaben' : 'Übersicht'} ebene="uebersicht">
	{#snippet aktionen()}<Export liste={reiter} {bereich} {art} anzahl={liste.length} />{/snippet}
</KopfLeiste>

<Ziehen aktualisieren={auffrischen}>
	<main class="mx-auto max-w-3xl px-4 pb-12">
		<!-- Primary Tabs direkt unter der Kopfleiste (M3): zwei Datensätze derselben Ansicht. -->
		<div class="-mx-4">
			<Reiter
				etikett="Meldungen"
				reiter={[
					{ id: 'offen', label: `Offen · ${zustand.offen.length}` },
					{ id: 'erledigt', label: 'Erledigt' }
				]}
				aktiv={reiter}
				onwahl={(w) => (reiter = w)}
			/>
		</div>
		<MailStoerung />

		<!-- Ein Filter mit nur einer Möglichkeit fehlt (M3), außer er ist gerade gewählt; ohne Filter
		     fehlt die Zeile. py-2: Die 48-px-Trefferflächen der Chips ragen 8 px über den Chip; so
		     bleiben sie unter den Reitern. -->
		{#if zeigeBereich || zeigeArt}
			<div class="flex min-h-14 flex-wrap content-center items-center gap-x-2 gap-y-4 py-2">
				{#if zeigeBereich}<MenueChip titel="Bereich" alle="Alle Bereiche" optionen={bereiche.map((b) => ({ wert: b, text: b }))} bind:wert={bereich} />{/if}
				{#if zeigeArt}<MenueChip titel="Art" alle="Alle Arten" optionen={arten} bind:wert={art} />{/if}
			</div>
		{:else}
			<div class="h-3"></div>
		{/if}

		{#if liste.length === 0 && quelle.length > 0}
			<div class="flex flex-col items-center px-6 py-16 text-center" in:fade={{ duration: 200 }}>
				<p class="text-title-large text-on-surface">Nichts in dieser Auswahl</p>
				<p class="mt-2 max-w-xs text-body-medium text-on-surface-variant">Die übrigen Meldungen stehen ohne Filter in der Liste.</p>
				<Button variant="tonal" class="mt-6" onclick={() => ((bereich = null), (art = null))}>Alle zeigen</Button>
			</div>
		{:else if liste.length === 0}
			<div class="flex flex-col items-center px-6 py-16 text-center" in:fade={{ duration: 200 }}>
				<span class="mb-5 grid size-20 place-items-center rounded-2xl bg-success-container text-on-success-container">
					<Icon name="task_alt" groesse={40} />
				</span>
				<p class="text-title-large text-on-surface">{reiter === 'offen' ? 'Alles erledigt' : 'Noch nichts erledigt'}</p>
				<p class="mt-2 max-w-xs text-body-medium text-on-surface-variant">
					{reiter === 'offen' ? 'Gerade ist kein Schaden offen.' : 'Erledigte Meldungen der letzten 30 Tage stehen hier.'}
				</p>
			</div>
		{:else}
			<!-- Segmentierte Liste (M3 Expressive): gefüllte Einträge mit 2 px Abstand, außen 16 px rund,
			     innen 4 px; je Abschnitt eine Liste unter ihrer Überschrift (M3 list subheader). -->
			<section class="flex flex-col gap-2" aria-label="Meldungen">
				{#each abschnitte as a (a.titel)}
					{#if a.titel}<h2 class="px-1 pt-2 text-title-small text-on-surface-variant">{a.titel}</h2>{/if}
					<ul class="flex flex-col gap-0.5">
						{#each a.meldungen as m (m.id)}
							<li class="rounded-xs bg-surface-container-low first:rounded-t-lg last:rounded-b-lg" animate:flip={{ duration: 300 }} out:slide={{ duration: 200 }}>
								<MeldungZeile meldung={m} />
							</li>
						{/each}
					</ul>
				{/each}
			</section>
		{/if}
	</main>
</Ziehen>
