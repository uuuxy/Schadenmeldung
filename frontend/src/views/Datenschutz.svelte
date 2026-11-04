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
				<li>wer welche Rolle hat (Hausmeister, Leitung, Verwaltung, Admin) und wer welche Mails und Mitteilungen bekommt</li>
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
			<h2 class="mb-2 text-title-medium">Mitteilungen aufs Handy</h2>
			<p class="text-on-surface-variant">
				Hausmeister, Schulleitung, Sekretariat und Admins können sie je Gerät einschalten. Dann kommt zu jeder neuen Meldung eine Mitteilung mit Schaden und Ort, ohne Namen, und am App-Symbol steht die Zahl der neuen Meldungen. Dafür speichert die App je Gerät die Adresse, über die der Dienst des Browser-Herstellers (Google, Apple, Mozilla oder Microsoft) zustellt. Der Inhalt ist verschlüsselt; der Dienst sieht nur, dass eine Nachricht an das Gerät geht. Die Adresse wird gelöscht beim Ausschalten, beim Abmelden und wenn die App 180 Tage nicht geöffnet wurde.
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Export und Weiterleiten</h2>
			<p class="text-on-surface-variant">
				Hausmeister und Leitung können eine einzelne Meldung über das Teilen-Menü ihres Geräts weiterleiten, etwa an den Schulträger: Ort, Schaden, Notiz und Foto, ohne den Namen der meldenden Person.
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
				Fotos werden {e.fotoTage} Tage nach der Erledigung gelöscht, die Meldung selbst nach {e.aufbewahrungTage} Tagen. Jede Nacht und vor jedem Update sichert der Server die Datenbank; jede Sicherung wird nach 7 Tagen gelöscht. Alles bleibt auf dem Server der Schule; es gibt keine Weitergabe und kein Tracking. Einen fremden Dienst nutzen nur die Mitteilungen aufs Handy, verschlüsselt und nur für alle, die sie einschalten.
			</p>
		</section>
		<section>
			<h2 class="mb-2 text-title-medium">Fragen</h2>
			<p class="text-on-surface-variant">Ansprechpartner ist die oder der Datenschutzbeauftragte der Schule.</p>
		</section>
	</div>
</main>
