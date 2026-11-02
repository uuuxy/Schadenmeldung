<script>
	/**
	 * @component Feld — das Eingabefeld nach Material 3 (outlined), wie ui/Feld der Bibliothek.
	 *
	 * Die Beschriftung steht über dem Feld statt schwebend: Ein schwebendes Label wandert beim Tippen
	 * weg und ist im Ruhezustand vom Platzhalter nicht zu unterscheiden.
	 *
	 * Rahmen `outline` (3:1 Kontrast für den Rand eines Bedienelements, WCAG 1.4.11), Fläche
	 * surface-container-lowest, im Fokus primary mit 1-px-Ring. Höhe 48 px fürs Handy (die
	 * Bibliothek nimmt 36 px am Rechner) und Schrift 16 px, sonst zoomt iOS beim Antippen.
	 *
	 * Der Hilfetext hängt an aria-describedby: Ein Name benennt, eine Beschreibung erklärt. Mit
	 * maxlength zeigt das Feld beim Tippen einen Zähler (M3: bei einer Grenze immer ein Zähler), sonst
	 * bliebe es an der Grenze einfach stumm stehen.
	 */
	let {
		wert = $bindable(''),
		etikett,
		typ = 'text',
		hinweis = '',
		mehrzeilig = false,
		zeilen = 3,
		element = $bindable(),
		nachlaufend,
		...rest
	} = $props();
	const id = $props.id();
	let fokus = $state(false);
	const grenze = $derived(Number(rest.maxlength) || 0);
	const klasse = $derived([
		'block w-full rounded-md border border-outline bg-surface-container-lowest px-4 text-body-large text-on-surface caret-primary transition-colors',
		'placeholder:text-outline focus:border-primary focus:ring-1 focus:ring-primary focus:outline-none',
		mehrzeilig ? 'resize-none py-3' : 'h-12',
		nachlaufend && 'pr-12'
	]);
</script>

<div class="flex flex-col gap-1.5">
	<label for={id} class="text-label-large text-on-surface-variant">{etikett}</label>
	<div class="relative">
		{#if mehrzeilig}
			<textarea {id} bind:value={wert} bind:this={element} rows={zeilen} aria-describedby={hinweis ? `${id}-hinweis` : undefined} class={klasse} onfocus={() => (fokus = true)} onblur={() => (fokus = false)} {...rest}></textarea>
		{:else}
			<input {id} type={typ} bind:value={wert} bind:this={element} aria-describedby={hinweis ? `${id}-hinweis` : undefined} class={klasse} onfocus={() => (fokus = true)} onblur={() => (fokus = false)} {...rest} />
		{/if}
		{#if nachlaufend}
			<div class="absolute top-1 right-1">{@render nachlaufend()}</div>
		{/if}
	</div>
	{#if hinweis || (grenze && fokus)}
		<div class="flex gap-4 text-body-small text-on-surface-variant">
			<p id="{id}-hinweis" class="flex-1">{hinweis}</p>
			{#if grenze && fokus}<span class="shrink-0 tabular-nums" aria-hidden="true">{(wert ?? '').length}/{grenze}</span>{/if}
		</div>
	{/if}
</div>
