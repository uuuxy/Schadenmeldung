<script>
	let { groesse = 24, class: klasse = '' } = $props();
</script>

<svg class={['ladekreis animate-spin', klasse]} width={groesse} height={groesse} viewBox="0 0 24 24" role="progressbar" aria-label="Lädt">
	<circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-dasharray="40 100" />
</svg>
