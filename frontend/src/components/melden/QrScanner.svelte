<script>
	import { onMount } from 'svelte';
	import Icon from '../Icon.svelte';
	import Ladekreis from '../Ladekreis.svelte';

	/**
	 * Den QR-Aufkleber in der App lesen. Die Kamera-App des iPhones öffnet ihn sonst in Safari, und dort
	 * gilt die Anmeldung der App vom Home-Bildschirm nicht. Gelesen wird auf dem Gerät (jsQR, erst beim
	 * Öffnen geladen), das Kamerabild verlässt es nicht, und zwar genau das Quadrat, das man sieht.
	 * onlesen bekommt den Text des Codes und gibt einen Hinweis zurück, wenn er zu keinem Raum führt;
	 * sonst schließt die Seite darüber den Scanner, und die Kamera geht aus.
	 */
	let { onlesen } = $props();

	const VERBOTEN = 'Kamera nicht erlaubt – in den Einstellungen freigeben';
	const KEINE = 'Keine Kamera gefunden';
	const BELEGT = 'Kamera wird gerade benutzt';
	const FEHLER = { NotAllowedError: VERBOTEN, SecurityError: VERBOTEN, NotFoundError: KEINE, OverconstrainedError: KEINE, NotReadableError: BELEGT, AbortError: BELEGT };

	let video = $state();
	let laeuft = $state(false);
	let fehler = $state('');
	let hinweis = $state('');

	onMount(() => {
		let strom = null;
		let uhr = 0;
		// Zählt Starts und Halte: Was ein überholter Start noch bekommt, hält er gleich wieder an.
		let lauf = 0;
		let lies = null;
		const leinwand = document.createElement('canvas');
		const pinsel = leinwand.getContext('2d', { willReadFrequently: true });

		async function starte() {
			const meiner = ++lauf;
			fehler = '';
			try {
				lies ??= (await import('jsqr')).default;
			} catch {
				if (meiner === lauf) fehler = 'Ohne Netz nicht verfügbar';
				return;
			}
			try {
				const neu = await navigator.mediaDevices.getUserMedia({ video: { facingMode: { ideal: 'environment' }, width: { ideal: 1280 }, height: { ideal: 720 } }, audio: false });
				if (meiner !== lauf) return neu.getTracks().forEach((t) => t.stop());
				strom = neu;
				video.srcObject = neu;
				await video.play();
				if (meiner !== lauf) return;
				laeuft = true;
				schaue(meiner);
			} catch (e) {
				if (meiner === lauf) fehler = FEHLER[e?.name] ?? 'Kamera startet nicht';
			}
		}

		function halte() {
			lauf++;
			clearTimeout(uhr);
			strom?.getTracks().forEach((t) => t.stop());
			strom = null;
			if (video) video.srcObject = null;
			laeuft = false;
		}

		// Etwa achtmal je Sekunde das sichtbare Quadrat lesen, in höchstens 480 px: Das genügt für einen
		// Aufkleber auf Armlänge und lässt schwächeren Handys Luft.
		function schaue(meiner) {
			if (meiner !== lauf) return;
			const { videoWidth: b, videoHeight: h } = video;
			if (b && h) {
				const seite = Math.min(b, h);
				const ziel = Math.min(seite, 480);
				if (leinwand.width !== ziel) leinwand.width = leinwand.height = ziel;
				pinsel.drawImage(video, (b - seite) / 2, (h - seite) / 2, seite, seite, 0, 0, ziel, ziel);
				const text = lies(pinsel.getImageData(0, 0, ziel, ziel).data, ziel, ziel, { inversionAttempts: 'dontInvert' })?.data;
				if (text) hinweis = onlesen(text) ?? '';
			}
			uhr = setTimeout(() => schaue(meiner), 120);
		}

		// Im Hintergrund hält das Handy die Kamera ohnehin an; zurück in der App startet sie neu, statt
		// schwarz stehen zu bleiben.
		const sicht = () => {
			halte();
			if (document.visibilityState === 'visible') starte();
		};
		document.addEventListener('visibilitychange', sicht);
		starte();
		return () => {
			document.removeEventListener('visibilitychange', sicht);
			halte();
		};
	});
</script>

<!-- Das Quadrat ist der Sucher: Gelesen wird genau, was darin zu sehen ist. -->
<div class="relative mx-auto grid aspect-square w-full max-w-sm place-items-center overflow-hidden rounded-xl bg-surface-container-highest">
	<!-- Ausgeschriebenes Schluss-Tag: <video /> ist kein void-Element. -->
	<video bind:this={video} autoplay playsinline muted aria-hidden="true" class="absolute inset-0 size-full object-cover"></video>
	{#if fehler}
		<div class="relative flex flex-col items-center gap-3 p-6 text-center">
			<Icon name="no_photography" groesse={32} class="text-on-surface-variant" />
			<p class="text-body-large text-on-surface" role="alert">{fehler}</p>
		</div>
	{:else if !laeuft}
		<div class="relative text-primary"><Ladekreis groesse={32} /></div>
	{/if}
</div>
<!-- Ohne Anleitung: Titel und Sucher sprechen für sich (M3). Nur ein Code, der nicht passt, sagt warum. -->
<p class="mt-3 rounded-md bg-warning-container px-4 py-3 text-body-medium text-on-warning-container empty:hidden" aria-live="polite">{fehler ? '' : hinweis}</p>
