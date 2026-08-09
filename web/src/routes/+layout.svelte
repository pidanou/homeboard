<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';

	const favicon = '/favicon.png';
	import { Toaster } from 'svelte-sonner';
	import { initAccentColor } from '$lib/theme';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const apply = (dark: boolean) => document.documentElement.classList.toggle('dark', dark);
		const onChange = (e: MediaQueryListEvent) => {
			if (!localStorage.getItem('theme')) apply(e.matches);
		};
		if (!localStorage.getItem('theme')) apply(mq.matches);
		mq.addEventListener('change', onChange);
		initAccentColor();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
<Toaster richColors theme="system" />
