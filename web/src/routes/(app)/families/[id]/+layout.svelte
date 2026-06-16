<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';

	let { children } = $props();

	type Family = { id: string; name: string };
	let family = $state<Family | null>(null);

	const familyID = $derived($page.params.id);

	onMount(async () => {
		family = await api.get<Family>(`/api/v1/families/${familyID}`);
	});
</script>

<!-- Family name shown on mobile (desktop sidebar already shows it) -->
{#if family}
	<h2 class="md:hidden text-lg font-semibold mb-4">{family.name}</h2>
{/if}

{@render children()}
