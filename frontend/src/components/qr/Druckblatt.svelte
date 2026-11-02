<script>
	/**
	 * Ein A4-Blatt in Originalgröße mit seinen Aufklebern. Die Maße kommen als CSS-Variablen vom Bogen
	 * (views/QrBogen.svelte). Am Bildschirm verkleinert zoom das Blatt samt Inhalt; der Rahmen liegt
	 * außen auf der Hülle, damit er nicht mitschrumpft. nurDruck blendet das Blatt am Bildschirm aus.
	 */
	let { raeume, format, qr, nurDruck = false } = $props();
</script>

<div class={['huelle overflow-hidden rounded-md border border-outline-variant', nurDruck && 'nur-druck']}>
	<div class={['seite bg-papier text-auf-papier', `anordnung-${format.anordnung}`, format.linien && 'mit-linien']}>
		{#each raeume as r (r.id)}
			<div class="etikett">
				<div class="qr">{@html qr(r)}</div>
				<div class="text">
					<p class="marke text-papier-akzent">Schaden melden</p>
					<p class="name">{r.name}</p>
					{#if r.bereich}<p class="bereich text-papier-leise">{r.bereich}</p>{/if}
					{#if format.schrift.tipp}<p class="tipp text-papier-leiser">Mit dem Handy scannen</p>{/if}
				</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.seite {
		zoom: var(--zoom);
		width: 210mm;
		height: 297mm;
		display: grid;
		grid-template-columns: repeat(var(--spalten), var(--breite));
		grid-auto-rows: var(--hoehe);
		align-content: start;
		padding: var(--oben) 0 0 var(--links);
		box-sizing: border-box;
	}
	.etikett {
		display: flex;
		align-items: center;
		min-width: 0;
		overflow: hidden;
		box-sizing: border-box;
		padding: var(--innen);
		outline: 0.25mm dashed var(--color-papier-linie);
		outline-offset: -0.25mm;
	}
	.anordnung-stapel .etikett {
		flex-direction: column;
		justify-content: center;
		text-align: center;
	}
	.anordnung-seite .etikett {
		gap: calc(var(--innen) * 0.6);
	}
	.qr {
		width: var(--qr);
		flex-shrink: 0;
	}
	.anordnung-stapel .qr {
		margin: calc(var(--innen) * 0.5) 0;
	}
	.qr :global(svg) {
		display: block;
		width: 100%;
		height: auto;
	}
	.text {
		min-width: 0;
	}
	.anordnung-seite .text {
		flex: 1;
	}
	.marke {
		font-size: var(--marke);
		font-weight: 500;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}
	.name {
		font-size: var(--name);
		font-weight: 500;
		line-height: 1.15;
		overflow-wrap: anywhere;
		hyphens: auto;
	}
	.bereich {
		font-size: var(--bereich);
		margin-top: 0.15em;
	}
	.tipp {
		font-size: var(--tipp);
		margin-top: 0.4em;
	}
	@media screen {
		.nur-druck {
			display: none;
		}
	}
	@media print {
		@page {
			size: A4;
			margin: 0;
		}
		.huelle {
			border: 0;
			border-radius: 0;
			break-after: page;
		}
		.huelle:last-child {
			break-after: auto;
		}
		.seite {
			zoom: 1;
		}
		/* Etikettenbögen sind schon geschnitten: gedruckt wird nur die Linie zum Ausschneiden. */
		.etikett {
			outline: none;
		}
		.mit-linien .etikett {
			outline: 0.2mm dashed var(--color-papier-linie);
		}
	}
</style>
