<script>
	import { zustand } from '../lib/zustand.svelte.js';
	import { zurueck } from '../lib/weg.svelte.js';
	import KopfLeiste from '../components/KopfLeiste.svelte';

	const e = $derived(zustand.einstellungen);
</script>

<KopfLeiste titel="Datenschutz" zurueck={() => zurueck('/')} konto={false} />

<main class="mx-auto max-w-2xl px-6 pt-2 pb-16 text-body-large text-on-surface">
	<div class="flex flex-col gap-6">
		<section>
			<h2 class="mb-2 text-title-medium">Wofür</h2>
			<p class="text-on-surface-variant">
				Mit dieser App meldet das Kollegium Schäden im Schulgebäude, und der Hausmeister zeigt den Stand, bis sie behoben sind.
				{#if e.schulname}Verantwortlich ist die Schule: {e.schulname}.{/if}
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Was gespeichert wird</h2>
			<ul class="list-disc space-y-1 pl-5 text-on-surface-variant">
				<li>Ort, Art und Beschreibung des Schadens, auf Wunsch ein Foto</li>
				<li>Zeitpunkt der Meldung und der Erledigung, der Stand und ein interner Vermerk, den das Kollegium nicht sieht</li>
				<li>wer gemeldet hat – sichtbar nur für Hausmeister, Schulleitung, Sekretariat und Admins, für Rückfragen</li>
				<li>wer welche Rolle hat (Hausmeister, Leitung, Verwaltung, Admin) und wer welche Mails bekommt</li>
			</ul>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Anmeldung</h2>
			<p class="text-on-surface-variant">
				Die Anmeldung läuft über das Postfach der Schule. Das Passwort wird nur an den Mailserver der Schule weitergereicht und nirgends gespeichert. Ein technisch notwendiges Cookie hält das Gerät angemeldet, höchstens ein Jahr lang.
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Mail an den Hausmeister</h2>
			<p class="text-on-surface-variant">
				Der Hausmeister bekommt jede neue Meldung sofort per Mail; das lässt sich in der App abschalten. Die Knöpfe in der Mail gelten 14 Tage und nur für die jeweilige Meldung. Angezeigt wird der Stand einer Meldung, nicht der Zeitpunkt, zu dem sie auf „In Arbeit“ gesetzt wurde.
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Export</h2>
			<p class="text-on-surface-variant">
				Hausmeister, Leitung und Verwaltung können die Meldungen als PDF oder Excel-Tabelle herunterladen, etwa für den Schulträger. Die Dateien enthalten keine Namen der Meldenden. Die Löschfristen gelten nur in der App: Wer eine Datei herunterlädt, löscht sie, sobald sie nicht mehr gebraucht wird.
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Fotos und Texte</h2>
			<p class="text-on-surface-variant">
				Bei Fotos werden Standort und Gerätedaten entfernt. Bitte keine Personen fotografieren und keine Namen von Schülerinnen und Schülern nennen.
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Löschung</h2>
			<p class="text-on-surface-variant">
				Fotos werden {e.fotoTage} Tage nach der Erledigung gelöscht, die Meldung selbst nach {e.aufbewahrungTage} Tagen. Jede Nacht und vor jedem Update sichert der Server die Datenbank; jede Sicherung wird nach 7 Tagen gelöscht. Alles bleibt auf dem Server der Schule; es gibt keine Weitergabe, keine fremden Dienste und kein Tracking.
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Fragen</h2>
			<p class="text-on-surface-variant">Ansprechpartner ist die oder der Datenschutzbeauftragte der Schule.</p>
		</section>
	</div>
</main>
