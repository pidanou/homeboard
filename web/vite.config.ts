import { paraglideVitePlugin } from '@inlang/paraglide-js';
import adapterNode from '@sveltejs/adapter-node';
import adapterStatic from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { VitePWA } from 'vite-plugin-pwa';

// ADAPTER=node builds the runtime-configurable server used by the Docker image.
// Default (static) is what Capacitor's `cap sync` bundles into the native shell.
const adapter = process.env.ADAPTER === 'node' ? adapterNode() : adapterStatic({ fallback: 'index.html' });

export default defineConfig({
	server: {
		allowedHosts: true,
		hmr: { host: process.env.LOCAL_IP || 'localhost' },
	},
	optimizeDeps: { exclude: ['svelte-sonner'] },
	plugins: [
		paraglideVitePlugin({
			project: './project.inlang',
			outdir: './src/lib/paraglide',
			strategy: ['localStorage', 'preferredLanguage', 'baseLocale']
		}),
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter,
			paths: { base: (process.env.BASE_PATH as `/${string}` | undefined) || '' }
		}),
		VitePWA({
			strategies: 'injectManifest',
			srcDir: 'src',
			filename: 'service-worker.ts',
			registerType: 'autoUpdate',
			manifest: {
				name: 'Homeboard',
				short_name: 'Homeboard',
				description: 'Your home, organised.',
				theme_color: '#9a6022',
				background_color: '#2b1f14',
				display: 'standalone',
				icons: [
					{ src: `${process.env.BASE_PATH || ''}/icon-192.png`, sizes: '192x192', type: 'image/png' },
					{ src: `${process.env.BASE_PATH || ''}/icon-512.png`, sizes: '512x512', type: 'image/png', purpose: 'any maskable' }
				]
			},
			injectManifest: {
				globPatterns: ['**/*.{js,css,html,ico,png,svg,webp}']
			},
			devOptions: {
				enabled: true,
				type: 'module'
			}
		})
	]
});
