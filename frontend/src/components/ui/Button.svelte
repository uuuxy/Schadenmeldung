<script>
	/**
	 * @component Button — der Knopf nach Material 3, gebaut wie ui/Button der Bibliothek.
	 *
	 * Eine Pille (M3, Shape: Buttons voll rund), ohne Schatten: In den M3-Token trägt ein Bauteil
	 * entweder einen Rahmen oder eine Erhebung, nie beides. Rückmeldung beim Zeigen und Drücken
	 * gibt der State-Layer aus styles/komponenten.css.
	 *
	 * Varianten nach M3, Buttons: `primary` gefüllt, `tonal` in secondary-container, `secondary`
	 * umrandet (outline-variant, on-surface-variant), `ghost` der Textknopf in primary. `success`
	 * und `danger` sind getönte Knöpfe in den Container-Rollen. `symbol` trägt nur ein Symbol und
	 * hat über .icon-btn die Trefferfläche von 48 px.
	 *
	 * Höhen nach M3 (Button-Größen): sm 32, md 40, lg 56 px. Am Handy ist 40 die Regel, 56 für die
	 * eine Hauptaktion einer Seite. sm und md treffen auf 48 px (.treffer in styles/komponenten.css).
	 */
	let { children, variant = 'primary', size = 'md', class: klasse = '', element = $bindable(), ...rest } = $props();

	const varianten = {
		primary: 'bg-primary text-on-primary border-transparent',
		tonal: 'bg-secondary-container text-on-secondary-container border-transparent',
		secondary: 'bg-surface-container-lowest border-outline-variant text-on-surface-variant',
		success: 'bg-success-container text-on-success-container border-transparent',
		danger: 'bg-error-container text-on-error-container border-transparent',
		ghost: 'bg-transparent border-transparent text-primary',
		symbol: 'icon-btn bg-transparent border-transparent text-on-surface-variant'
	};
	const groessen = {
		sm: 'h-8 px-3 gap-1.5 text-label-large',
		md: 'h-10 px-4 gap-2 text-label-large',
		lg: 'h-14 px-6 gap-2 text-title-medium'
	};
</script>

<!-- Gesperrt: ein Zustand für alle Varianten, nach M3 – Fläche on-surface 12 %, Beschriftung 38 %. -->
<button
	bind:this={element}
	class={[
		'inline-flex shrink-0 items-center justify-center rounded-full border transition-colors duration-200',
		'disabled:cursor-not-allowed disabled:border-transparent! disabled:bg-on-surface/12! disabled:text-on-surface/38!',
		variant !== 'symbol' && groessen[size],
		variant !== 'symbol' && size !== 'lg' && 'treffer',
		varianten[variant],
		klasse
	]}
	{...rest}
>
	{@render children?.()}
</button>
