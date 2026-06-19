<script lang="ts">
	import { onMount } from 'svelte';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from 'svelte-sonner';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const apply = (dark: boolean) => document.documentElement.classList.toggle('dark', dark);
		apply(mq.matches);
		mq.addEventListener('change', (e) => apply(e.matches));
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
<Toaster richColors />
