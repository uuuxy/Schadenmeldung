<script>
	import Icon from '../Icon.svelte';

	/**
	 * @component AssistChip — M3 Assist-Chip mit Symbol, 32 px hoch mit 48-px-Trefferfläche (.treffer).
	 * Ungewählt mit Rand, das Symbol in primary (bei gefahr in error); als Schalter gewählt gefüllt in
	 * secondary-container, bei gefahr in error-container. Mit onFoto wird er zur Fotowahl: ein Label um
	 * ein unsichtbares Dateifeld, am Handy öffnet das die Kamera.
	 *
	 * @prop {string} icon
	 * @prop {string} text
	 * @prop {boolean} [gewaehlt] Schalter: aria-pressed
	 * @prop {boolean} [gefahr]
	 * @prop {string} [iconGewaehlt] Symbol im gewählten Zustand; sonst wird icon gefüllt
	 * @prop {() => void} [onclick]
	 * @prop {(e: Event) => void} [onFoto]
	 */
	let { icon, text, gewaehlt = false, gefahr = false, iconGewaehlt = null, onclick, onFoto = null } = $props();
	const flaeche = $derived.by(() => {
		if (!gewaehlt) return 'border-outline-variant text-on-surface';
		return gefahr ? 'border-transparent bg-error-container text-on-error-container' : 'border-transparent bg-secondary-container text-on-secondary-container';
	});
	const klasse = $derived(['treffer flex h-8 items-center gap-2 rounded-sm border pr-4 pl-2 text-label-large transition-colors', flaeche]);
	const symbolFarbe = $derived(gewaehlt ? '' : gefahr ? 'text-error' : 'text-primary');
</script>

{#if onFoto}
	<label class={['schicht cursor-pointer', klasse]}>
		<Icon name={icon} groesse={18} class={symbolFarbe} />{text}
		<input type="file" accept="image/*" capture="environment" class="sr-only" onchange={onFoto} />
	</label>
{:else}
	<button class={klasse} aria-pressed={gewaehlt} {onclick}>
		<Icon name={gewaehlt && iconGewaehlt ? iconGewaehlt : icon} groesse={18} gefuellt={gewaehlt && !iconGewaehlt} class={symbolFarbe} />{text}
	</button>
{/if}
