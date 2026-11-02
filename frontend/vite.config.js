import { writeFileSync } from 'node:fs';
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

// Go bettet frontend/dist ein und braucht den Ordner schon vor dem ersten Bau; vite leert ihn beim
// Bauen, deshalb kommt der Platzhalter danach zurück.
function platzhalterBehalten() {
	return {
		name: 'platzhalter-behalten',
		closeBundle() {
			writeFileSync('dist/.gitkeep', '');
		}
	};
}

export default defineConfig({
	plugins: [tailwindcss(), svelte(), platzhalterBehalten()],
	server: {
		// host: true, damit das Handy im selben WLAN den Entwicklungsserver erreicht.
		host: true,
		proxy: {
			'/api': 'http://127.0.0.1:8090',
			'/fotos': 'http://127.0.0.1:8090'
		}
	},
	build: {
		target: 'es2022'
	}
});
