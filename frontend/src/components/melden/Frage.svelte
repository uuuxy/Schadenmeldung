<script>
	import Icon from '../Icon.svelte';

	/**
	 * Eine Frage des Meldeformulars. Beantwortet trägt die Überschrift ein Häkchen; tippt jemand zu
	 * früh auf „Melden“, wird sie rot und sagt, was fehlt – wie eine Pflichtfrage in Google Forms,
	 * statt dass der Knopf es ständig wiederholt.
	 */
	let { titel, erledigt = false, hinweis = '', element = $bindable(), children } = $props();
</script>

<section bind:this={element} class="scroll-mt-20">
	<h2 class={['mb-3 flex items-center gap-2 text-title-medium transition-colors', hinweis ? 'text-error' : 'text-on-surface']}>
		{titel}
		{#if erledigt}<Icon name="check_circle" groesse={20} gefuellt class="text-primary" />{/if}
	</h2>
	{#if hinweis}
		<p class="-mt-1.5 mb-3 flex items-center gap-1.5 text-body-medium text-error" role="alert">
			<Icon name="error" groesse={18} />{hinweis}
		</p>
	{/if}
	{@render children?.()}
</section>
