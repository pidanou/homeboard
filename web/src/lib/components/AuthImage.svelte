<script lang="ts">
	import { getBaseUrl } from '$lib/api/client';

	let {
		src,
		alt = '',
		class: className = ''
	}: {
		src?: string | null;
		alt?: string;
		class?: string;
	} = $props();

	const resolvedUrl = $derived((() => {
		if (!src) return null;
		try {
			return `${getBaseUrl()}${new URL(src).pathname}`;
		} catch {
			return src.startsWith('/') ? `${getBaseUrl()}${src}` : src;
		}
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
		return () => {
			active = false;
			if (localBlob) URL.revokeObjectURL(localBlob);
		};
	});
</script>

{#if blobUrl}
	<img src={blobUrl} {alt} class={className} />
{/if}
