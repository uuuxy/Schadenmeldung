<script>
	import { renderSVG } from 'uqr';
	import { zustand, darfBearbeiten } from '../lib/zustand.svelte.js';
	import { zurueck } from '../lib/weg.svelte.js';
	import Icon from '../components/Icon.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Button from '../components/ui/Button.svelte';

	// Der Code führt auf die Adresse, unter der die App gerade läuft.
	const basis = location.origin;
	// Dieselbe Reihenfolge wie in der Raumliste (vom Server): so liegen die Aufkleber eines Flurs beisammen.
	const raeume = $derived(zustand.raeume);
	const qr = (r) => renderSVG(`${basis}/r/${r.code}`, { border: 1, ecc: 'M' });
</script>

<div class="contents print:hidden">
	<KopfLeiste titel="QR-Aufkleber" konto={false} zurueck={() => zurueck(darfBearbeiten() ? '/einstellungen' : '/')}>
		{#snippet aktionen()}
			<Button class="mr-2" onclick={() => window.print()}><Icon name="print" groesse={18} />Drucken</Button>
		{/snippet}
	</KopfLeiste>
	<p class="mx-auto max-w-5xl px-4 pb-4 text-body-medium text-on-surface-variant">
		{raeume.length === 1 ? '1 Aufkleber' : `${raeume.length} Aufkleber`}, 12 je A4-Blatt. Der Code führt mit dem Handy direkt zur Meldung für diesen Raum.
	</p>
</div>

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
