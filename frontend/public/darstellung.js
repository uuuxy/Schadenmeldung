// Entscheidet vor dem ersten Bild, ob die App hell oder dunkel erscheint – dieselbe Regel wie
// src/lib/darstellung.svelte.js: die Wahl auf diesem Gerät, sonst das System. Eigene Datei, weil
// die Sicherheitsregeln (CSP) keine eingebetteten Skripte erlauben.
(function () {
	try {
		var wahl = localStorage.getItem('schaden.darstellung');
		var dunkel = wahl === 'dunkel' || (wahl !== 'hell' && matchMedia('(prefers-color-scheme: dark)').matches);
		document.documentElement.classList.toggle('schema-dunkel', dunkel);
	} catch (e) {
		/* Ohne Speicher bleibt es beim hellen Schema. */
	}
})();
