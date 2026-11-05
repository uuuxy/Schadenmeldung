<script>
	/**
	 * @component Button — der Knopf nach Material 3, gebaut wie ui/Button der Bibliothek.
	 *
	 * Eine Pille (M3, Shape: Buttons voll rund), ohne Schatten: In den M3-Token trägt ein Bauteil
	 * entweder einen Rahmen oder eine Erhebung, nie beides. Rückmeldung beim Zeigen und Drücken
	 * gibt der State-Layer aus styles/komponenten.css. Gedrückt wird die Pille eckiger und federt
	 * zurück (M3 Expressive, Buttons: Shape morph, gedrückt 8 dp, ab 56 px Höhe 12 dp); damit sich die
	 * Ecke stufenlos ändert, ist die Rundung die halbe Höhe statt „voll“. Nach dem Loslassen bleibt sie
	 * 150 ms eckig, sonst läge der Formwechsel nur unter dem Finger.
	 *
	 * Varianten nach M3, Buttons: `primary` gefüllt, `tonal` in secondary-container, `secondary`
	 * umrandet (outline-variant, on-surface-variant), `ghost` der Textknopf in primary, `warnend` der
	 * Textknopf für Endgültiges in error (wie „Löschen“ in Googles Dialogen). `symbol` trägt nur ein Symbol und
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
		ghost: 'bg-transparent border-transparent text-primary',
		warnend: 'bg-transparent border-transparent text-error',
		symbol: 'icon-btn bg-transparent border-transparent text-on-surface-variant'
	};
	const groessen = {
		sm: 'h-8 px-3 gap-1.5 text-label-large rounded-lg enabled:active:rounded-sm',
		md: 'h-10 px-4 gap-2 text-label-large rounded-xl enabled:active:rounded-sm',
		lg: 'h-14 px-6 gap-2 text-title-medium rounded-2xl enabled:active:rounded-md'
	};
</script>

<!-- Gesperrt: ein Zustand für alle Varianten, nach M3 – Fläche on-surface 12 %, Beschriftung 38 %.
     aria-disabled sieht genauso aus, bleibt aber antippbar und fokussierbar („soft-disabled“ in
     material-web): So kann ein Knopf beim Antippen sagen, was noch fehlt. -->
<button
	bind:this={element}
	class={[
		// Die Ecke federt (fast spatial), Farben gehen ohne Überschwingen über (M3: effects).
		'inline-flex shrink-0 items-center justify-center border enabled:active:delay-0 [transition:border-radius_350ms_var(--ease-spring)_150ms,color_200ms_var(--ease-standard),background-color_200ms_var(--ease-standard),border-color_200ms_var(--ease-standard)]',
		'disabled:cursor-not-allowed disabled:border-transparent! disabled:bg-on-surface/12! disabled:text-on-surface/38!',
		'aria-disabled:border-transparent! aria-disabled:bg-on-surface/12! aria-disabled:text-on-surface/38!',
		variant !== 'symbol' && groessen[size],
		variant !== 'symbol' && size !== 'lg' && 'treffer',
		varianten[variant],
		klasse
	]}
	{...rest}
>
	{@render children?.()}
</button>
