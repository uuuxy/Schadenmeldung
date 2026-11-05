<script>
	import Icon from './Icon.svelte';

	/**
	 * Zurück aus einer Unteransicht wie aus einem Untermenü: vorn der Pfeil, daneben, wo man gerade ist
	 * (M3: der Zurück-Pfeil steht vorn und führt eine Ebene zurück). Ein Bereich, „Weitere Räume“ und
	 * „Anderer Ort“ nutzen dieselbe Zeile, damit „zurück“ überall gleich aussieht und an derselben Stelle steht.
	 */
	let { titel, etikett, onzurueck } = $props();
</script>

<button class="mt-3 -ml-2 flex min-h-12 items-center gap-1 rounded-full pr-4 text-on-surface" aria-label={etikett} onclick={onzurueck}>
	<span class="grid size-10 place-items-center"><Icon name="arrow_back" /></span>
	<span class="text-title-medium">{titel}</span>
</button>
