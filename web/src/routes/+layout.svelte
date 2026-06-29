<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isLocal } from '$lib/env';

	const favicon = '/favicon.png';
	import { Toaster } from 'svelte-sonner';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const apply = (dark: boolean) => document.documentElement.classList.toggle('dark', dark);
		apply(mq.matches);
		mq.addEventListener('change', (e) => apply(e.matches));

		const isNative = !!(window as any).Capacitor?.isNativePlatform?.();
		if (isLocal && isNative && !localStorage.getItem('api_url') && !import.meta.env.PUBLIC_API_URL && $page.url.pathname !== '/setup') {
			goto('/setup');
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
<Toaster richColors theme="system" />
