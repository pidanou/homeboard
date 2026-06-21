<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { env } from '$env/dynamic/public';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from 'svelte-sonner';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const apply = (dark: boolean) => document.documentElement.classList.toggle('dark', dark);
		apply(mq.matches);
		mq.addEventListener('change', (e) => apply(e.matches));

		if (!env.PUBLIC_API_URL && !localStorage.getItem('api_url') && $page.url.pathname !== '/setup') {
			goto('/setup');
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
<Toaster richColors theme="system" />
