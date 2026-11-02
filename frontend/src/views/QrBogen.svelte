<script>
	import { renderSVG } from 'uqr';
	import { zustand } from '../lib/zustand.svelte.js';
	import { gehe } from '../lib/weg.svelte.js';
	import Icon from '../components/Icon.svelte';

	// Der Code führt auf die Adresse, unter der die App gerade läuft – also vom Server aus drucken.
	const basis = location.origin;
	// Dieselbe Reihenfolge wie in der Raumliste (vom Server): so liegen die Aufkleber eines Flurs beisammen.
	const raeume = $derived(zustand.raeume);
	const qr = (r) => renderSVG(`${basis}/r/${r.code}`, { border: 1, ecc: 'M' });
</script>

<div class="sticky top-0 z-10 flex items-center gap-2 bg-surface px-2 pt-[env(safe-area-inset-top)] print:hidden">
	<div class="mx-auto flex h-16 w-full max-w-5xl items-center gap-2">
		<button class="icon-btn text-on-surface" aria-label="Zurück" onclick={() => gehe('/verwaltung')}><Icon name="arrow_back" /></button>
		<h1 class="flex-1 text-title-large text-on-surface">QR-Aufkleber</h1>
		<button class="inline-flex h-10 items-center gap-2 rounded-full bg-primary pr-6 pl-4 text-label-large text-on-primary" onclick={() => window.print()}>
			<Icon name="print" groesse={18} /> Drucken
		</button>
	</div>
</div>

<p class="mx-auto max-w-5xl px-4 pb-4 text-body-medium text-on-surface-variant print:hidden">
	Codes führen auf {basis}. Zum Ausdrucken die Seite vom Server aus öffnen, nicht vom Entwicklungsrechner.
</p>

<div class="bogen mx-auto grid max-w-5xl grid-cols-2 gap-3 px-4 pb-10 sm:grid-cols-3 print:max-w-none print:gap-0 print:p-0">
	{#each raeume as r (r.id)}
		<div class="etikett flex flex-col items-center justify-center rounded-xl bg-white p-4 text-center text-black print:rounded-none">
			<p class="text-[11px] font-medium tracking-wide text-[#0061a4] uppercase">Schaden melden</p>
			<div class="qr my-2 w-full max-w-40">{@html qr(r)}</div>
			<p class="text-xl leading-tight font-medium">{r.name}</p>
			{#if r.bereich}<p class="text-xs text-neutral-600">{r.bereich}</p>{/if}
			<p class="mt-1 text-[10px] text-neutral-500">Mit dem Handy scannen</p>
		</div>
	{/each}
</div>

<style>
	.qr :global(svg) {
		display: block;
		width: 100%;
		height: auto;
	}
	@media print {
		@page {
			size: A4;
			margin: 10mm;
		}
		.bogen {
			grid-template-columns: repeat(3, 1fr);
		}
		.etikett {
			height: 69mm;
			break-inside: avoid;
			border: 0.2mm dashed #bbb;
		}
		.qr {
			width: 36mm;
		}
	}
</style>
