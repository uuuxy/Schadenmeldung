<script>
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';

	/**
	 * Filter „Bereich“ als ein Chip mit Pfeil statt einer Reihe von Chips: M3 rät ab mehr als zwei
	 * Reihen davon ab, und ein Menü wird am Handy zum Blatt. Gewählt trägt der Chip den Bereich und
	 * liegt auf secondary-container, wie jeder gewählte Filter-Chip.
	 */
	let { bereiche, bereich = $bindable() } = $props();
	let offen = $state(false);

	function waehle(b) {
		bereich = b;
		offen = false;
	}
</script>

<button
	class={[
		'treffer flex h-8 min-w-0 items-center gap-2 rounded-sm border pr-2 text-label-large transition-colors duration-200',
		bereich ? 'border-transparent bg-secondary-container pl-2 text-on-secondary-container' : 'border-outline pl-4 text-on-surface-variant'
	]}
	aria-haspopup="dialog"
	onclick={() => (offen = true)}
>
	{#if bereich}<Icon name="check" groesse={18} />{/if}
	<span class="truncate">{bereich ?? 'Bereich'}</span>
	<Icon name="arrow_drop_down" groesse={18} />
</button>

<Blatt bind:offen titel="Bereich">
	<ul class="flex flex-col gap-1 px-3 pb-4">
		{#each [null, ...bereiche] as b (b ?? '')}
			<li>
				<button
					class={['flex h-14 w-full items-center gap-3 rounded-full px-4 text-left text-body-large', b === bereich ? 'bg-secondary-container text-on-secondary-container' : 'text-on-surface']}
					aria-pressed={b === bereich}
					onclick={() => waehle(b)}
				>
					<Icon name="check" class={b === bereich ? '' : 'invisible'} />
					<span class="truncate">{b ?? 'Alle Bereiche'}</span>
				</button>
			</li>
		{/each}
	</ul>
</Blatt>
