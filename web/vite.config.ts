import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
	server: { allowedHosts: true },
	optimizeDeps: { exclude: ['svelte-sonner'] },
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({ fallback: 'index.html' })
		}),
		VitePWA({
			registerType: 'autoUpdate',
			manifest: {
				name: 'Homeboard',
				short_name: 'Homeboard',
				description: 'Your home, organised.',
				theme_color: '#9a6022',
				background_color: '#2b1f14',
				display: 'standalone',
				icons: [
					{ src: '/icon.svg', sizes: 'any', type: 'image/svg+xml', purpose: 'any maskable' }
				]
			},
			workbox: {
				globPatterns: ['**/*.{js,css,html,ico,png,svg,webp}']
			}
		})
	]
});
