<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { X, Plus, Pencil, CornerDownLeft, Trash2 } from 'lucide-svelte';
	import type { AppList, AppListItem } from '$lib/types';

	const familyID = $derived($page.params.id ?? '');

	let lists = $state<AppList[]>([]);
	let itemsByList = $state<Record<string, AppListItem[]>>({});
	let newItemByList = $state<Record<string, string>>({});
	let addingList = $state(false);
	let newListName = $state('');
	let confirmDeleteListID = $state<string | null>(null);
	let renamingListID = $state<string | null>(null);
	let renameListValue = $state('');
	let renamingItem = $state<{ listID: string; itemID: string } | null>(null);
	let renameItemValue = $state('');
	let activeListID = $state<string | null>(null);

	let es: EventSource | null = null;

	function focusSelect(el: HTMLElement) { (el as HTMLInputElement).focus(); (el as HTMLInputElement).select(); }

	async function loadAll() {
		const res = await api.get<AppList[]>(`/api/v1/households/${familyID}/lists`);
		lists = res ?? [];
		if (!activeListID && lists.length > 0) activeListID = lists[0].id;
		const results = await Promise.allSettled(
			lists.map(l => api.get<AppListItem[]>(`/api/v1/households/${familyID}/lists/${l.id}/items`))
		);
		const map: Record<string, AppListItem[]> = {};
		lists.forEach((l, i) => {
			const r = results[i];
			map[l.id] = r.status === 'fulfilled' ? (r.value ?? []) : (itemsByList[l.id] ?? []);
		});
		itemsByList = map;
	}

	onMount(() => {
		loadAll();
		es = new EventSource(sseUrl(`/api/v1/households/${familyID}/stream`) as string);
		es.onmessage = (e) => { if (e.data === 'refresh') loadAll(); };
		es.onerror = () => { es?.close(); es = null; };
	});

	onDestroy(() => es?.close());

	async function createList() {
		if (!newListName.trim()) return;
		try {
			const list = await api.post<AppList>(`/api/v1/households/${familyID}/lists`, { name: newListName.trim() });
			lists = [...lists, list];
			itemsByList = { ...itemsByList, [list.id]: [] };
			activeListID = list.id;
			newListName = '';
			addingList = false;
		} catch {}
	}

	async function deleteList(id: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/lists/${id}`);
			lists = lists.filter(l => l.id !== id);
			const { [id]: _, ...rest } = itemsByList;
			itemsByList = rest;
			if (activeListID === id) activeListID = lists[0]?.id ?? null;
			confirmDeleteListID = null;
		} catch {}
	}

	async function addItem(listID: string) {
		const name = (newItemByList[listID] ?? '').trim();
		if (!name) return;
		try {
			const item = await api.post<AppListItem>(`/api/v1/households/${familyID}/lists/${listID}/items`, { name });
			itemsByList = { ...itemsByList, [listID]: [item, ...(itemsByList[listID] ?? [])] };
			newItemByList = { ...newItemByList, [listID]: '' };
		} catch {}
	}

	async function toggleItem(listID: string, item: AppListItem) {
		try {
			await api.patch(`/api/v1/households/${familyID}/lists/${listID}/items/${item.id}`, {
				name: item.name, checked: !item.checked,
			});
			itemsByList = {
				...itemsByList,
				[listID]: (itemsByList[listID] ?? []).map(i => i.id === item.id ? { ...i, checked: !item.checked } : i),
			};
		} catch {}
	}

	async function deleteItem(listID: string, itemID: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/lists/${listID}/items/${itemID}`);
			itemsByList = { ...itemsByList, [listID]: (itemsByList[listID] ?? []).filter(i => i.id !== itemID) };
		} catch {}
	}

	async function clearChecked(listID: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/lists/${listID}/items/checked`);
			itemsByList = { ...itemsByList, [listID]: (itemsByList[listID] ?? []).filter(i => !i.checked) };
		} catch {}
	}

	async function submitRenameList() {
		if (!renamingListID || !renameListValue.trim()) { renamingListID = null; return; }
		try {
			await api.patch(`/api/v1/households/${familyID}/lists/${renamingListID}`, { name: renameListValue.trim() });
			lists = lists.map(l => l.id === renamingListID ? { ...l, name: renameListValue.trim() } : l);
		} catch {}
		renamingListID = null;
	}

	async function submitRenameItem() {
		if (!renamingItem || !renameItemValue.trim()) { renamingItem = null; return; }
		const { listID, itemID } = renamingItem;
		const item = (itemsByList[listID] ?? []).find(i => i.id === itemID);
		if (!item) { renamingItem = null; return; }
		try {
			await api.patch(`/api/v1/households/${familyID}/lists/${listID}/items/${itemID}`, { name: renameItemValue.trim(), checked: item.checked });
			itemsByList = {
				...itemsByList,
				[listID]: (itemsByList[listID] ?? []).map(i => i.id === itemID ? { ...i, name: renameItemValue.trim() } : i),
			};
		} catch {}
		renamingItem = null;
	}

	function unchecked(listID: string) {
		return (itemsByList[listID] ?? [])
			.filter(i => !i.checked)
			.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
	}

	function checked(listID: string) {
		return (itemsByList[listID] ?? [])
			.filter(i => i.checked)
			.sort((a, b) => new Date(b.checked_at ?? b.created_at).getTime() - new Date(a.checked_at ?? a.created_at).getTime());
	}
</script>

<!-- ===================== MOBILE ===================== -->
<div class="md:hidden flex flex-col">

	<!-- Sticky header + tabs -->
	<div class="sticky top-0 z-10 bg-background px-4 pt-4 pb-2 flex flex-col gap-2">

		<!-- Title row — always visible -->
		<div class="flex items-center justify-between">
			<h1 class="text-xl font-semibold">Lists {#if lists.length > 0}<span class="text-base font-normal text-muted-foreground ml-1">{lists.length}</span>{/if}</h1>
			<button onclick={() => (addingList = true)} class="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors" aria-label="New list">
				<Plus class="w-4 h-4" />
			</button>
		</div>

		<!-- Add list form -->
		{#if addingList}
			<div class="flex items-center gap-2">
				<Input
					bind:value={newListName}
					placeholder="List name…"
					class="flex-1 h-8 text-sm"
					autofocus
					onkeydown={(e) => {
						if (e.key === 'Enter') { e.preventDefault(); createList(); }
						if (e.key === 'Escape') { addingList = false; newListName = ''; }
					}}
				/>
				<Button size="sm" class="h-8" onclick={createList} disabled={!newListName.trim()}>Add</Button>
				<Button size="sm" variant="ghost" class="h-8" onclick={() => { addingList = false; newListName = ''; }}>Cancel</Button>
			</div>
		{/if}

		<!-- Delete confirmation -->
		{#if confirmDeleteListID}
			{@const target = lists.find(l => l.id === confirmDeleteListID)}
			<div class="flex items-center gap-2">
				<span class="flex-1 text-sm text-muted-foreground truncate">Delete <strong>{target?.name}</strong>?</span>
				<Button size="sm" variant="destructive" onclick={() => deleteList(confirmDeleteListID!)} class="h-8">Delete</Button>
				<Button size="sm" variant="ghost" onclick={() => (confirmDeleteListID = null)} class="h-8">Cancel</Button>
			</div>
		{/if}

		<!-- Tabs -->
		{#if lists.length > 0}
			<div class="flex items-center gap-1.5 overflow-x-auto py-1 -my-1">
				{#each lists as list (list.id)}
					{#if activeListID === list.id}
						<div class="flex items-center gap-1 pl-3 pr-1.5 py-1.5 rounded-full bg-primary text-primary-foreground shrink-0">
							<span class="text-sm font-medium whitespace-nowrap">{list.name}</span>
							<button
								onclick={() => (confirmDeleteListID = list.id)}
								class="flex items-center justify-center w-5 h-5 rounded-full hover:bg-primary-foreground/20 transition-colors"
								aria-label="Delete list"
							>
								<X class="w-3 h-3" />
							</button>
						</div>
					{:else}
						<button
							onclick={() => (activeListID = list.id)}
							class="px-3 py-1.5 rounded-full text-sm font-medium whitespace-nowrap bg-muted text-muted-foreground hover:bg-muted/80 transition-colors shrink-0"
						>{list.name}</button>
					{/if}
				{/each}
			</div>
		{/if}
	</div>

	<!-- Active list content -->
	{#if activeListID}
		{@const lid = activeListID}
		{@const uncheckedItems = unchecked(lid)}
		{@const checkedItems = checked(lid)}

		<!-- Add item -->
		<div class="flex items-center gap-2 px-4 py-3 border-b border-border">
			<Plus class="w-3.5 h-3.5 text-muted-foreground/50 shrink-0" />
			<input
				class="flex-1 text-sm bg-transparent outline-none placeholder:text-muted-foreground/50"
				placeholder="Add item…"
				bind:value={newItemByList[lid]}
				onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addItem(lid); } }}
			/>
			<button
				onclick={() => addItem(lid)}
				disabled={!(newItemByList[lid] ?? '').trim()}
				class="p-1 rounded-md text-muted-foreground/50 hover:text-foreground hover:bg-muted transition-colors shrink-0 disabled:pointer-events-none"
				aria-label="Add item"
			>
				<CornerDownLeft class="w-3.5 h-3.5" />
			</button>
		</div>

		<!-- Items -->
		{#if uncheckedItems.length === 0 && checkedItems.length === 0}
			<p class="text-sm text-muted-foreground text-center py-16 italic">Empty — add something above</p>
		{:else}
			<div class="flex flex-col divide-y divide-border pb-8">
				{#each uncheckedItems as item (item.id)}
					<div class="flex items-center gap-3 px-4 py-3">
						<Checkbox checked={false} onCheckedChange={() => toggleItem(lid, item)} />
						<span class="flex-1 text-sm">{item.name}</span>
						<button onclick={() => deleteItem(lid, item.id)} class="p-1 text-muted-foreground hover:text-destructive transition-colors shrink-0" aria-label="Delete">
							<X class="w-3.5 h-3.5" />
						</button>
					</div>
				{/each}

				{#if checkedItems.length > 0}
					<div class="flex items-center justify-between px-4 py-2 bg-muted/40">
						<span class="text-xs font-medium text-muted-foreground">In cart ({checkedItems.length})</span>
						<button onclick={() => clearChecked(lid)} class="text-xs text-muted-foreground hover:text-foreground transition-colors">Clear all</button>
					</div>
					{#each checkedItems as item (item.id)}
						<div class="flex items-center gap-3 px-4 py-3 opacity-50">
							<Checkbox checked={true} onCheckedChange={() => toggleItem(lid, item)} />
							<span class="flex-1 text-sm line-through">{item.name}</span>
							<button onclick={() => deleteItem(lid, item.id)} class="p-1 text-muted-foreground hover:text-destructive transition-colors shrink-0" aria-label="Delete">
								<X class="w-3.5 h-3.5" />
							</button>
						</div>
					{/each}
				{/if}
			</div>
		{/if}
	{:else if !addingList}
		<div class="flex flex-col items-center gap-2 py-16 px-4 text-muted-foreground">
			<p class="text-sm font-medium">No lists yet</p>
			<p class="text-xs">Tap + to create one.</p>
		</div>
	{/if}
</div>

<!-- ===================== DESKTOP ===================== -->
<div class="hidden md:block">

	<!-- Header -->
	<div class="sticky top-0 z-10 bg-background px-6 pt-6 pb-3">
		<h1 class="text-xl font-semibold">Lists {#if lists.length > 0}<span class="text-base font-normal text-muted-foreground ml-1">{lists.length}</span>{/if}</h1>
	</div>

	<!-- Kanban board -->
	<div class="relative">
		<div class="overflow-x-auto px-6 pb-8 pt-3">
		<div class="flex gap-4 items-start" style="min-width: max-content">

			<!-- Ghost "New list" column -->
			{#if addingList}
				<div class="w-72 flex flex-col rounded-xl border border-border bg-card overflow-hidden shrink-0">
					<div class="h-14 flex items-center px-4 border-b border-border bg-muted/50">
						<span class="text-sm font-semibold text-muted-foreground">New list</span>
					</div>
					<div class="p-4 flex flex-col gap-2">
						<Input
							bind:value={newListName}
							placeholder="List name…"
							class="h-8 text-sm"
							autofocus
							onkeydown={(e) => {
								if (e.key === 'Enter') { e.preventDefault(); createList(); }
								if (e.key === 'Escape') { addingList = false; newListName = ''; }
							}}
						/>
						<div class="flex gap-2">
							<Button size="sm" class="flex-1" onclick={createList} disabled={!newListName.trim()}>Create</Button>
							<Button size="sm" variant="ghost" onclick={() => { addingList = false; newListName = ''; }}>Cancel</Button>
						</div>
					</div>
				</div>
			{:else}
				<button
					onclick={() => (addingList = true)}
					class="w-72 shrink-0 flex flex-col items-center justify-center gap-1.5 h-32 rounded-xl border-2 border-dashed border-border text-muted-foreground hover:border-primary/50 hover:text-primary transition-colors"
				>
					<Plus class="w-5 h-5" />
					<span class="text-sm font-medium">New list</span>
				</button>
			{/if}

			{#each lists as list (list.id)}
				{@const uncheckedItems = unchecked(list.id)}
				{@const checkedItems = checked(list.id)}
				<div class="w-72 flex flex-col rounded-xl border border-border bg-card overflow-hidden shrink-0">

					<!-- Column header -->
					<div class="flex items-center gap-1 px-4 h-14 border-b border-border bg-muted/50">
						{#if renamingListID === list.id}
							<input
								class="flex-1 text-sm font-semibold bg-transparent border-none outline-none border-b border-primary"
								bind:value={renameListValue}
								onblur={submitRenameList}
								onkeydown={(e) => { if (e.key === 'Enter') submitRenameList(); if (e.key === 'Escape') renamingListID = null; }}
								use:focusSelect
							/>
						{:else}
							<span class="flex-1 text-sm font-semibold truncate">{list.name}</span>
							{#if confirmDeleteListID === list.id}
								<span class="text-xs text-muted-foreground shrink-0">Delete?</span>
								<Button size="sm" variant="destructive" onclick={() => deleteList(list.id)} class="h-6 px-2 text-xs shrink-0">Yes</Button>
								<button onclick={() => (confirmDeleteListID = null)} class="p-1 text-muted-foreground hover:text-foreground shrink-0" aria-label="Cancel">
									<X class="w-3 h-3" />
								</button>
							{:else}
								<button
									onclick={() => { renamingListID = list.id; renameListValue = list.name; }}
									class="p-1 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors shrink-0"
									aria-label="Rename list"
								>
									<Pencil class="w-3.5 h-3.5" />
								</button>
								<button
									onclick={() => (confirmDeleteListID = list.id)}
									class="p-1 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors shrink-0"
									aria-label="Delete list"
								>
									<Trash2 class="w-3.5 h-3.5" />
								</button>
							{/if}
						{/if}
					</div>

					<!-- Add item -->
					<div class="flex items-center gap-2 px-4 py-2.5 border-b border-border">
						<Plus class="w-3.5 h-3.5 text-muted-foreground/50 shrink-0" />
						<input
							class="flex-1 text-sm bg-transparent outline-none placeholder:text-muted-foreground/50"
							placeholder="Add item…"
							bind:value={newItemByList[list.id]}
							onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addItem(list.id); } }}
						/>
						<button
							onclick={() => addItem(list.id)}
							disabled={!(newItemByList[list.id] ?? '').trim()}
							class="p-1 rounded-md text-muted-foreground/50 hover:text-foreground hover:bg-muted transition-colors shrink-0 disabled:pointer-events-none"
							aria-label="Add item"
						>
							<CornerDownLeft class="w-3.5 h-3.5" />
						</button>
					</div>

					<!-- Items -->
					{#if uncheckedItems.length === 0 && checkedItems.length === 0}
						<p class="text-xs text-muted-foreground text-center py-8 italic">Empty</p>
					{:else}
						<div class="flex flex-col divide-y divide-border">
							{#each uncheckedItems as item (item.id)}
								<div class="flex items-center gap-3 px-4 py-2.5">
									<Checkbox checked={false} onCheckedChange={() => toggleItem(list.id, item)} />
									{#if renamingItem?.listID === list.id && renamingItem?.itemID === item.id}
										<input
											class="flex-1 text-sm bg-transparent border-none outline-none border-b border-primary"
											bind:value={renameItemValue}
											onblur={submitRenameItem}
											onkeydown={(e) => { if (e.key === 'Enter') submitRenameItem(); if (e.key === 'Escape') renamingItem = null; }}
											use:focusSelect
										/>
									{:else}
										<button
											class="flex-1 text-sm text-left"
											ondblclick={() => { renamingItem = { listID: list.id, itemID: item.id }; renameItemValue = item.name; }}
										>{item.name}</button>
									{/if}
									<button onclick={() => deleteItem(list.id, item.id)} class="p-1 text-muted-foreground hover:text-destructive transition-colors shrink-0" aria-label="Delete">
										<X class="w-3.5 h-3.5" />
									</button>
								</div>
							{/each}

							{#if checkedItems.length > 0}
								<div class="flex items-center justify-between px-4 py-2 bg-muted/40">
									<span class="text-xs font-medium text-muted-foreground">In cart ({checkedItems.length})</span>
									<button onclick={() => clearChecked(list.id)} class="text-xs text-muted-foreground hover:text-foreground transition-colors">
										Clear all
									</button>
								</div>
								{#each checkedItems as item (item.id)}
									<div class="flex items-center gap-3 px-4 py-2.5 opacity-50">
										<Checkbox checked={true} onCheckedChange={() => toggleItem(list.id, item)} />
										<span class="flex-1 text-sm line-through">{item.name}</span>
										<button onclick={() => deleteItem(list.id, item.id)} class="p-1 text-muted-foreground hover:text-destructive transition-colors shrink-0" aria-label="Delete">
											<X class="w-3.5 h-3.5" />
										</button>
									</div>
								{/each}
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>
		</div>
		<div class="pointer-events-none absolute inset-y-0 right-0 w-12 bg-gradient-to-l from-background to-transparent"></div>
	</div>
</div>
