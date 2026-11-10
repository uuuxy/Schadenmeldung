<script>
	import { slide } from 'svelte/transition';
	import { verkleinereFoto } from '../../lib/foto.js';
	import Icon from '../Icon.svelte';
	import AssistChip from '../ui/AssistChip.svelte';
	import Feld from '../ui/Feld.svelte';
	import FilterChips from '../ui/FilterChips.svelte';

	/**
	 * Was genau? Die typischen Fälle der Art als Chips, „Anderes …“ öffnet ein kurzes Feld. Darunter,
	 * ausdrücklich freiwillig: Foto, Notiz, Dringend (Assist-Chips nach M3, Symbol in primary).
	 */
	let { art, freiBeschreiben, fall = $bindable(), notiz = $bindable(), dringend = $bindable(), foto = $bindable() } = $props();
	let notizOffen = $state(false);
	let fotoUrl = $state('');

	// Ein neues Foto ersetzt das alte; wird es weggenommen (auch nach dem Melden), fällt die Vorschau weg.
	$effect(() => {
		if (!foto) return;
		const url = URL.createObjectURL(foto);
		fotoUrl = url;
		return () => {
			URL.revokeObjectURL(url);
			fotoUrl = '';
		};
	});

	async function fotoGewaehlt(e) {
		const datei = e.currentTarget.files?.[0];
		e.currentTarget.value = '';
		if (!datei) return;
		try {
			foto = await verkleinereFoto(datei);
		} catch {
			foto = datei; // Kann der Browser das Format nicht umrechnen, prüft es der Server.
		}
	}
</script>

{#if art.faelle.length}
	<FilterChips
		etikett="Was genau"
		optionen={[...art.faelle.map((f) => ({ wert: f, text: f })), { wert: '__anderes', text: 'Anderes …' }]}
		wert={fall || null}
		onwahl={(f) => (fall = f ?? '')}
	/>
{/if}
{#if freiBeschreiben}
	<div class={art.faelle.length && 'mt-4'} transition:slide={{ duration: 200 }}>
		<Feld etikett="Was ist kaputt?" bind:wert={notiz} maxlength="160" enterkeyhint="done" autocomplete="off" />
	</div>
{/if}

<p class="mt-6 mb-2 text-label-large text-on-surface-variant">Freiwillig</p>
<div class="flex flex-wrap items-center gap-x-2 gap-y-4">
	<AssistChip icon="photo_camera" text={foto ? 'Anderes Foto' : 'Foto'} onFoto={fotoGewaehlt} />
	{#if !freiBeschreiben}
		<AssistChip icon="edit_note" iconGewaehlt="check" text="Notiz" gewaehlt={notizOffen} onclick={() => (notizOffen = !notizOffen)} />
	{/if}
	<AssistChip icon="warning" text="Dringend / Gefahr" gefahr gewaehlt={dringend} onclick={() => (dringend = !dringend)} />
</div>

{#if fotoUrl}
	<div class="relative mt-4 inline-block" transition:slide={{ duration: 200 }}>
		<img src={fotoUrl} alt="Gewähltes Foto" class="h-28 w-28 rounded-md object-cover" />
		<button class="icon-btn absolute -top-2 -right-2 size-8 bg-inverse-surface text-inverse-on-surface" aria-label="Foto entfernen" onclick={() => (foto = null)}>
			<Icon name="close" groesse={18} />
		</button>
	</div>
{/if}

{#if notizOffen && !freiBeschreiben}
	<div class="mt-4" transition:slide={{ duration: 200 }}>
		<Feld etikett="Notiz" bind:wert={notiz} mehrzeilig zeilen={2} maxlength="600" hinweis="Bitte keine Namen von Schülerinnen und Schülern." />
	</div>
{/if}
