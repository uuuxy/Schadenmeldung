<script>
	import { symbole } from '../lib/symbole.js';

	/** Material Symbol (Rounded). `gefuellt` zeigt die volle Form für Gewähltes, wie M3 es vorsieht. */
	let { name, gefuellt = false, groesse = 24, class: klasse = '' } = $props();
	const d = $derived(symbole[gefuellt && symbole[`${name}-fill`] ? `${name}-fill` : name] ?? '');
</script>

<svg class={['shrink-0 fill-current', klasse]} width={groesse} height={groesse} viewBox="0 -960 960 960" aria-hidden="true">
	<path {d} />
</svg>
