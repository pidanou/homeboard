<script lang="ts">
	import { base } from '$app/paths';
	import { api, sseUrl } from '$lib/api/client';
	import type { AppList, AppListItem } from '$lib/types';
	import * as m from '$lib/paraglide/messages.js';

	let { familyID }: { familyID: string } = $props();

	let list = $state<AppList | null>(null);
	let items = $state<AppListItem[]>([]);
	let es: EventSource | null = null;

	async function load() {
		const lists = await api.get<AppList[]>(`/api/v1/households/${familyID}/lists`);
		const first = [...(lists ?? [])].sort((a, b) => a.position - b.position)[0] ?? null;
		list = first;
		items = first ? ((await api.get<AppListItem[]>(`/api/v1/households/${familyID}/lists/${first.id}/items`)) ?? []) : [];
	}

	$effect(() => {
		load();
		es?.close();
		es = new EventSource(sseUrl(`/api/v1/households/${familyID}/stream`));
		es.onmessage = (e) => { if (e.data === 'refresh') load(); };
		es.onerror = () => { es?.close(); es = null; };
		return () => es?.close();
	});

	const unchecked = $derived(
		items
			.filter((i) => !i.checked)
			.sort((a, b) => {
				if (a.manual_order != null && b.manual_order != null) return a.manual_order - b.manual_order;
				if (a.manual_order != null) return -1;
				if (b.manual_order != null) return 1;
				return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
			}),
	);
</script>

<div>
	<div class="flex items-center gap-3 mb-3">
		<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-foreground">{m.today_lists_label()}</span>
		<div class="flex-1 h-px bg-border"></div>
		{#if list}
			<a href="{base}/households/{familyID}/lists" class="text-xs text-muted-foreground hover:text-foreground shrink-0 transition-colors">
				{m.today_lists_view_all()}
			</a>
		{/if}
	</div>
	{#if list}
		<div class="rounded-lg border border-border bg-card px-4 py-3">
			<div class="flex items-center justify-between gap-2 mb-1.5">
				<p class="text-sm font-medium truncate">{list.name}</p>
				{#if unchecked.length > 0}
					<span class="text-xs font-medium text-muted-foreground bg-muted rounded-full px-1.5 shrink-0">{m.today_lists_remaining({ count: unchecked.length })}</span>
				{/if}
			</div>
			{#if unchecked.length > 0}
				<ul class="flex flex-col gap-1">
					{#each unchecked.slice(0, 3) as item (item.id)}
						<li class="text-xs text-muted-foreground truncate">{item.name}</li>
					{/each}
				</ul>
			{:else}
				<p class="text-xs text-muted-foreground/50 italic">{m.today_lists_empty()}</p>
			{/if}
		</div>
	{:else}
		<p class="text-sm text-muted-foreground/50 italic">{m.today_lists_empty()}</p>
	{/if}
</div>
