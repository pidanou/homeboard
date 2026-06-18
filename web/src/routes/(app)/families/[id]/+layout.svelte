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
	<p class="md:hidden text-xs text-muted-foreground font-medium uppercase tracking-wide mb-2 pt-4 px-4 md:px-6">{family.name}</p>
{/if}

{@render children()}
