<script lang="ts">
	import { getBaseUrl } from '$lib/api/client';
	import { onDestroy } from 'svelte';

	let {
		name,
		photoUrl,
		size = 32,
		class: className = ''
	}: {
		name: string;
		photoUrl?: string | null;
		size?: number;
		class?: string;
	} = $props();

	const resolvedUrl = $derived((() => {
		if (!photoUrl) return null;
		try { return `${getBaseUrl()}${new URL(photoUrl).pathname}`; }
		catch { return photoUrl.startsWith('/') ? `${getBaseUrl()}${photoUrl}` : photoUrl; }
	})());

	let blobUrl = $state<string | null>(null);

	$effect(() => {
		const url = resolvedUrl;
		if (!url) { blobUrl = null; return; }
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		let localBlob: string | null = null;
		let active = true;
		fetch(url, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
			.then(r => r.ok ? r.blob() : null)
			.then(blob => {
				if (!active) return;
				localBlob = blob ? URL.createObjectURL(blob) : null;
				blobUrl = localBlob;
			})
			.catch(() => { if (active) blobUrl = null; });
		return () => { active = false; if (localBlob) URL.revokeObjectURL(localBlob); };
	});

	onDestroy(() => { if (blobUrl) URL.revokeObjectURL(blobUrl); });

	const initials = $derived(name.trim().split(/\s+/).map(w => w[0]).join('').slice(0, 2).toUpperCase());

	const hue = $derived((() => {
		let h = 0;
		for (let i = 0; i < name.length; i++) h += name.charCodeAt(i);
		return h % 360;
	})());
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
		style="width:{size}px;height:{size}px;background:hsl({hue},55%,65%);font-size:{Math.round(size*0.38)}px"
		class="rounded-full flex items-center justify-center font-semibold text-white shrink-0 select-none {className}"
		aria-label={name}
	>
		{initials}
	</div>
{/if}
