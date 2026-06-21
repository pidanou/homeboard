<script lang="ts">
	import { getBaseUrl } from '$lib/api/client';
	import { onDestroy } from 'svelte';

	let {
		name,
		avatarUrl,
		userId,
		size = 40,
		class: className = ''
	}: {
		name: string;
		avatarUrl?: string | null;
		userId?: string;
		size?: number;
		class?: string;
	} = $props();

	const resolvedUrl = $derived((() => {
		if (!avatarUrl) return null;
		try {
			return `${getBaseUrl()}${new URL(avatarUrl).pathname}`;
		} catch {
			return avatarUrl.startsWith('/') ? `${getBaseUrl()}${avatarUrl}` : avatarUrl;
		}
	})());

	let blobUrl = $state<string | null>(null);
	let prevFetched = '';

	$effect(() => {
		const url = resolvedUrl;
		if (!url || url === prevFetched) return;
		prevFetched = url;
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		fetch(url, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
			.then(r => r.ok ? r.blob() : null)
			.then(blob => {
				if (blobUrl) URL.revokeObjectURL(blobUrl);
				blobUrl = blob ? URL.createObjectURL(blob) : null;
			})
			.catch(() => { blobUrl = null; });
	});

	onDestroy(() => { if (blobUrl) URL.revokeObjectURL(blobUrl); });

	const initials = $derived(() => {
		const parts = name.trim().split(/\s+/);
		if (parts.length === 1) return parts[0][0]?.toUpperCase() ?? '?';
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	});

	// ponytail: deterministic hue from userId or name so color is stable across renders
	const hue = $derived(() => {
		const seed = userId ?? name;
		let h = 0;
		for (let i = 0; i < seed.length; i++) h += seed.charCodeAt(i);
		return h % 360;
	});

	const bg = $derived(`hsl(${hue()}, 55%, 65%)`);
</script>

{#if blobUrl}
	<img
		src={blobUrl}
		alt={name}
		style="width:{size}px;height:{size}px"
		class="rounded-full object-cover shrink-0 {className}"
	/>
{:else}
	<div
		style="width:{size}px;height:{size}px;background:{bg};font-size:{Math.round(size * 0.38)}px"
		class="rounded-full flex items-center justify-center font-semibold text-white shrink-0 select-none {className}"
		aria-label={name}
	>
		{initials()}
	</div>
{/if}
