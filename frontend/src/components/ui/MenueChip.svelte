<script>
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';

	/**
	 * @component MenueChip — ein Filter-Chip mit Pfeil, der seine Auswahl als Blatt öffnet.
	 *
	 * M3 (Chips): Mehr als zwei Reihen Filter-Chips sind zu viel, dann trägt ein Chip ein Menü. Am
	 * Handy öffnet der ganze Chip das Menü, weil der Pfeil allein zu klein zum Treffen wäre, und das
	 * Menü wird zum Bottom Sheet. Gewählt trägt der Chip die Wahl mit Häkchen auf secondary-container
	 * wie jeder gewählte Filter-Chip; die erste Zeile des Blatts („Alle …“) nimmt die Wahl zurück.
	 * Optionen mit Symbol zeigen es im runden Feld wie die Zeilen der Liste, gewählt mit Häkchen hinten.
	 *
	 * @prop {string} titel - Name des Filters: auf dem Chip, solange nichts gewählt ist, und über dem Blatt.
	 * @prop {string} alle - die Zeile, die die Wahl zurücknimmt, z. B. „Alle Bereiche“.
	 * @prop {{ wert: string, text: string, kurz?: string, icon?: string, symbol?: string, flaeche?: string }[]} optionen
	 *   kurz: Beschriftung auf dem Chip, wenn der volle Text zu lang ist; symbol und flaeche: Farbklassen.
	 * @prop {string | null} wert - gebunden; null heißt alle.
	 */
	let { titel, alle, optionen, wert = $bindable(null) } = $props();
	let offen = $state(false);

	const gewaehlt = $derived(optionen.find((o) => o.wert === wert) ?? null);
	const mitSymbol = $derived(optionen.some((o) => o.icon));
	const zeilen = $derived([{ wert: null, text: alle, icon: 'category', flaeche: 'bg-surface-container-highest text-on-surface-variant' }, ...optionen]);

	function waehle(w) {
		wert = w;
		offen = false;
	}
</script>

<button
	class={[
		'treffer flex h-8 min-w-0 items-center gap-2 rounded-sm border pr-2 text-label-large transition-colors duration-200',
		gewaehlt ? 'border-transparent bg-secondary-container pl-2 text-on-secondary-container' : 'border-outline pl-4 text-on-surface-variant'
	]}
	aria-haspopup="dialog"
	aria-label={gewaehlt ? `${titel}: ${gewaehlt.text}` : titel}
	onclick={() => (offen = true)}
>
	{#if gewaehlt}<Icon name="check" groesse={18} />{/if}
	<span class="truncate">{gewaehlt ? (gewaehlt.kurz ?? gewaehlt.text) : titel}</span>
	<Icon name="arrow_drop_down" groesse={18} />
</button>

<Blatt bind:offen {titel}>
	<ul class="flex flex-col gap-1 px-3 pb-4">
		{#each zeilen as o (o.wert ?? '')}
			{@const an = o.wert === (gewaehlt?.wert ?? null)}
			<li>
				<button
					class={['flex h-14 w-full items-center gap-3 rounded-full px-4 text-left text-body-large', an ? 'bg-secondary-container text-on-secondary-container' : 'text-on-surface', mitSymbol && 'pl-2']}
					aria-pressed={an}
					onclick={() => waehle(o.wert)}
				>
					{#if mitSymbol}
						<span class={['grid size-10 shrink-0 place-items-center rounded-full', o.flaeche]}>
							<Icon name={o.icon} groesse={22} gefuellt={an} class={o.symbol} />
						</span>
						<span class="min-w-0 flex-1 truncate">{o.text}</span>
						<Icon name="check" class={an ? '' : 'invisible'} />
					{:else}
						<Icon name="check" class={an ? '' : 'invisible'} />
						<span class="truncate">{o.text}</span>
					{/if}
				</button>
			</li>
		{/each}
	</ul>
</Blatt>
