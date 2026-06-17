<script lang="ts">
	import type { Member, AppLabel, Filter } from '$lib/types';
	import { chipClass, dotClass } from '$lib/labels';
	import { Tag } from 'lucide-svelte';

	let {
		filter = $bindable<Filter>('all'),
		filterMembers = $bindable(new Set<string>()),
		filterLabels = $bindable(new Set<string>()),
		sortBy = $bindable<'date' | 'priority' | 'title'>('date'),
		sortAsc = $bindable(true),
		members,
		labels,
		doneCnt,
		familyID,
	}: {
		filter?: Filter;
		filterMembers?: Set<string>;
		filterLabels?: Set<string>;
		sortBy?: 'date' | 'priority' | 'title';
		sortAsc?: boolean;
		members: Member[];
		labels: AppLabel[];
		doneCnt: number;
		familyID: string;
	} = $props();

	let filterOpen = $state(false);
	let sortOpen = $state(false);

	const activeFilterCount = $derived(
		(filter !== 'all' ? 1 : 0) + filterMembers.size + filterLabels.size,
	);
	const sortChanged = $derived(sortBy !== 'date' || !sortAsc);

	const FILTERS: { id: Filter; label: string }[] = $derived([
		{ id: 'all', label: 'Active' },
		{ id: 'tasks', label: 'Tasks' },
		{ id: 'events', label: 'Events' },
		{ id: 'done', label: `Done${doneCnt ? ` (${doneCnt})` : ''}` },
	]);
</script>

<div class="flex items-center gap-2">
	<!-- Filter -->
	<div class="relative">
		{#if filterOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div class="fixed inset-0 z-10" onclick={() => (filterOpen = false)}></div>
		{/if}
		<button
			onclick={() => { filterOpen = !filterOpen; sortOpen = false; }}
			class="flex items-center gap-1.5 px-3 py-1.5 rounded-md border text-sm font-medium transition-colors cursor-pointer
				{activeFilterCount > 0
					? 'border-primary bg-primary/5 text-primary'
					: 'border-border bg-background text-muted-foreground hover:bg-muted/50'}"
		>
			<svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
				<path d="M2 4h12M4 8h8M6 12h4" />
			</svg>
			Filter
			{#if activeFilterCount > 0}
				<span class="ml-0.5 inline-flex items-center justify-center w-4 h-4 rounded-full bg-primary text-primary-foreground text-[10px] font-bold">
					{activeFilterCount}
				</span>
			{/if}
		</button>
		{#if filterOpen}
			<div class="absolute top-full left-0 mt-1.5 z-20 bg-popover border border-border rounded-lg shadow-lg p-4 min-w-56 flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<p class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Show</p>
					<div class="flex flex-wrap gap-1.5">
						{#each FILTERS as f}
							<button
								onclick={() => (filter = f.id)}
								class="px-2.5 py-1 rounded-full text-xs font-medium transition-colors cursor-pointer
									{filter === f.id ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground hover:bg-muted/80'}"
							>{f.label}</button>
						{/each}
					</div>
				</div>
				{#if members.length > 0}
					<div class="flex flex-col gap-2">
						<p class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Who</p>
						<div class="flex flex-wrap gap-1.5">
							<button
								onclick={() => (filterMembers = new Set())}
								class="px-2.5 py-1 rounded-full text-xs font-medium transition-colors cursor-pointer
									{filterMembers.size === 0 ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground hover:bg-muted/80'}"
							>Everyone</button>
							{#each members as m}
								<button
									onclick={() => {
										const next = new Set(filterMembers);
										next.has(m.user_id) ? next.delete(m.user_id) : next.add(m.user_id);
										filterMembers = next;
									}}
									class="px-2.5 py-1 rounded-full text-xs font-medium transition-colors cursor-pointer
										{filterMembers.has(m.user_id) ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground hover:bg-muted/80'}"
								>{m.name}</button>
							{/each}
						</div>
					</div>
				{/if}
				{#if labels.length > 0}
					<div class="flex flex-col gap-2">
						<p class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Labels</p>
						<div class="flex flex-wrap gap-1.5">
							{#each labels as lbl}
								<button
									onclick={() => {
										const next = new Set(filterLabels);
										next.has(lbl.id) ? next.delete(lbl.id) : next.add(lbl.id);
										filterLabels = next;
									}}
									class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-colors cursor-pointer
										{filterLabels.has(lbl.id) ? 'bg-foreground text-background' : chipClass(lbl.color)}"
								>
									<span class="w-1.5 h-1.5 rounded-full {dotClass(lbl.color)} shrink-0"></span>
									{lbl.name}
								</button>
							{/each}
						</div>
					</div>
				{/if}
				{#if activeFilterCount > 0}
					<button
						onclick={() => { filter = 'all'; filterMembers = new Set(); filterLabels = new Set(); }}
						class="text-xs text-muted-foreground hover:text-foreground text-left transition-colors cursor-pointer"
					>Clear all</button>
				{/if}
				<a
					href="/families/{familyID}/settings"
					onclick={() => (filterOpen = false)}
					class="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors pt-1 border-t border-border"
				>
					<Tag class="w-3 h-3" />
					Manage labels
				</a>
			</div>
		{/if}
	</div>

	<!-- Sort -->
	<div class="relative">
		{#if sortOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div class="fixed inset-0 z-10" onclick={() => (sortOpen = false)}></div>
		{/if}
		<button
			onclick={() => { sortOpen = !sortOpen; filterOpen = false; }}
			class="flex items-center gap-1.5 px-3 py-1.5 rounded-md border text-sm font-medium transition-colors cursor-pointer
				{sortChanged
					? 'border-primary bg-primary/5 text-primary'
					: 'border-border bg-background text-muted-foreground hover:bg-muted/50'}"
		>
			<svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
				<path d="M2 4h8M2 8h5M2 12h3M11 4v8M9 10l2 2 2-2" />
			</svg>
			Sort
		</button>
		{#if sortOpen}
			<div class="absolute top-full left-0 mt-1.5 z-20 bg-popover border border-border rounded-lg shadow-lg p-4 min-w-44 flex flex-col gap-3">
				<div class="flex flex-col gap-2">
					<p class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Order by</p>
					<div class="flex flex-col gap-1">
						{#each [['date', 'Date'], ['priority', 'Priority'], ['title', 'Title']] as const as [val, label]}
							<button
								onclick={() => (sortBy = val)}
								class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors text-left cursor-pointer
									{sortBy === val ? 'bg-muted font-medium' : 'text-muted-foreground hover:bg-muted/50'}"
							>
								<span class="w-3 h-3 rounded-full border-2 inline-block {sortBy === val ? 'border-foreground bg-foreground' : 'border-muted-foreground'}"></span>
								{label}
							</button>
						{/each}
					</div>
				</div>
				<div class="flex flex-col gap-1 border-t border-border pt-3">
					<button
						onclick={() => (sortAsc = true)}
						class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors text-left cursor-pointer
							{sortAsc ? 'bg-muted font-medium' : 'text-muted-foreground hover:bg-muted/50'}"
					>
						<span class="w-3 h-3 rounded-full border-2 inline-block {sortAsc ? 'border-foreground bg-foreground' : 'border-muted-foreground'}"></span>
						Ascending
					</button>
					<button
						onclick={() => (sortAsc = false)}
						class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors text-left cursor-pointer
							{!sortAsc ? 'bg-muted font-medium' : 'text-muted-foreground hover:bg-muted/50'}"
					>
						<span class="w-3 h-3 rounded-full border-2 inline-block {!sortAsc ? 'border-foreground bg-foreground' : 'border-muted-foreground'}"></span>
						Descending
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
