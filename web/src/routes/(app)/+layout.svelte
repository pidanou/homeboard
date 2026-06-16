<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { isLoggedIn, logout } from '$lib/auth';

	let { children } = $props();
	let ready = $state(false);

	onMount(() => {
		if (!isLoggedIn()) {
			goto('/login');
		} else {
			ready = true;
		}
	});
</script>

{#if ready}
	<div class="min-h-screen flex flex-col">
		<header class="border-b px-4 py-3 flex items-center justify-between">
			<a href="/" class="font-semibold text-lg">Family Board</a>
			<button onclick={logout} class="text-sm text-muted-foreground hover:underline">Sign out</button>
		</header>
		<main class="flex-1 p-4">
			{@render children()}
		</main>
	</div>
{/if}
