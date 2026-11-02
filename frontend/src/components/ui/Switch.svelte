<!-- @component Switch — der Ein/Aus-Schalter nach Material 3, übernommen aus ui/Switch der Bibliothek.

     Ein Häkchen wählt aus einer Liste aus, ein Schalter legt einen Zustand um („Dunkles Design“,
     „Mail bei dringenden Meldungen“).

     Maße nach M3 (Switch, Specs): Spur 52×32 dp, Griff 16 dp aus, 24 dp an und 28 dp beim Drücken,
     Zustandsschicht 40 dp, Trefferfläche 48 dp. Das Wachsen des Griffs macht den Zustand auch ohne
     Farbe erkennbar. Farben nach den Token (material-web, md-comp-switch): an Spur primary, Griff
     on-primary (beim Zeigen primary-container); aus Spur surface-container-highest mit Rand outline,
     Griff outline (beim Zeigen on-surface-variant). Der Griff hat keinen Schatten.

     Der Aufrufer gibt eine Beschriftung mit: sichtbar über `id` und ein eigenes <label>, oder per
     `etikett` als aria-label. -->
<script>
	let { checked = $bindable(false), etikett = '', id, disabled = false, onchange } = $props();

	const spur = $derived(
		disabled
			? checked
				? 'border-transparent bg-on-surface/12'
				: 'border-on-surface/12 bg-surface-container-highest/12'
			: checked
				? 'border-primary bg-primary'
				: 'border-outline bg-surface-container-highest'
	);
	const griff = $derived(
		disabled
			? checked
				? 'size-6 bg-surface'
				: 'size-4 bg-on-surface/38'
			: checked
				? 'size-6 bg-on-primary group-hover:bg-primary-container group-focus-visible:bg-primary-container group-active:size-7 group-active:bg-primary-container'
				: 'size-4 bg-outline group-hover:bg-on-surface-variant group-focus-visible:bg-on-surface-variant group-active:size-7 group-active:bg-on-surface-variant'
	);

	function umlegen() {
		if (disabled) return;
		checked = !checked;
		onchange?.(checked);
	}
</script>

<!-- before: die Trefferfläche, 52 × 48 px; sie ragt über den Knopf hinaus und verschiebt nichts. -->
<button
	{id}
	type="button"
	role="switch"
	aria-checked={checked}
	aria-label={etikett || undefined}
	{disabled}
	onclick={umlegen}
	class="group ohne-schicht relative inline-flex h-8 w-13 shrink-0 rounded-full border-2 transition-colors duration-200 before:absolute before:-inset-x-0.5 before:-inset-y-2.5 focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed {spur} {disabled
		? ''
		: 'cursor-pointer'}"
>
	<span class="pointer-events-none absolute -top-0.5 -left-0.5 flex size-8 items-center justify-center transition-transform duration-200 ease-in-out {checked ? 'translate-x-5' : ''}">
		<span
			class="absolute top-1/2 left-1/2 size-10 -translate-x-1/2 -translate-y-1/2 rounded-full opacity-0 transition-opacity group-hover:opacity-8 group-focus-visible:opacity-10 group-active:opacity-10 group-disabled:hidden {checked
				? 'bg-primary'
				: 'bg-on-surface'}"
		></span>
		<span class="relative rounded-full transition-all duration-200 ease-in-out {griff}"></span>
	</span>
</button>
