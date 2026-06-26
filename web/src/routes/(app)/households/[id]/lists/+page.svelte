<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { X, ShoppingCart } from 'lucide-svelte';
	import type { AppList, AppListItem } from '$lib/types';

	const familyID = $derived($page.params.id ?? '');

	const STORAGE_KEY = $derived(`activeListID:${familyID}`);

	let lists = $state<AppList[]>([]);
	let items = $state<AppListItem[]>([]);
	let activeListID = $state('');
	let newItemName = $state('');
	let addingList = $state(false);
	let newListName = $state('');
	let confirmDeleteList = $state<AppList | null>(null);
	function focusSelect(el: HTMLElement) { (el as HTMLInputElement).focus(); (el as HTMLInputElement).select(); }
	let renamingListID = $state<string | null>(null);
	let renameListValue = $state('');
	let renamingItemID = $state<string | null>(null);
	let renameItemValue = $state('');

	let es: EventSource | null = null;

	async function loadLists() {
		const res = await api.get<AppList[]>(`/api/v1/households/${familyID}/lists`);
		lists = res ?? [];
		const saved = localStorage.getItem(STORAGE_KEY);
		if (saved && lists.find((l) => l.id === saved)) {
			activeListID = saved;
		} else if (lists.length > 0 && !lists.find((l) => l.id === activeListID)) {
			activeListID = lists[0].id;
		}
	}

	async function loadItems() {
		if (!activeListID) return;
		const res = await api.get<AppListItem[]>(`/api/v1/households/${familyID}/lists/${activeListID}/items`);
		items = res ?? [];
	}

	async function loadAll() {
		await loadLists();
		await loadItems();
	}

	$effect(() => {
		if (activeListID) {
			localStorage.setItem(STORAGE_KEY, activeListID);
			loadItems();
		}
	});

	onMount(() => {
		loadAll();
		es = new EventSource(sseUrl(`/api/v1/households/${familyID}/stream`) as string);
		es.onmessage = (e) => { if (e.data === 'refresh') loadAll(); };
		es.onerror = () => { es?.close(); es = null; };
	});

	onDestroy(() => es?.close());

	async function createList(e: SubmitEvent) {
		e.preventDefault();
		if (!newListName.trim()) return;
		try {
			const list = await api.post<AppList>(`/api/v1/households/${familyID}/lists`, { name: newListName.trim() });
			lists = [...lists, list];
			activeListID = list.id;
			newListName = '';
			addingList = false;
		} catch { }
	}

	async function deleteList(id: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/lists/${id}`);
			lists = lists.filter((l) => l.id !== id);
			confirmDeleteList = null;
			if (activeListID === id) {
				activeListID = lists[0]?.id ?? '';
				items = [];
			}
		} catch { }
	}

	async function addItem(e: SubmitEvent) {
		e.preventDefault();
		if (!newItemName.trim() || !activeListID) return;
		try {
			const item = await api.post<AppListItem>(`/api/v1/households/${familyID}/lists/${activeListID}/items`, { name: newItemName.trim() });
			items = [item, ...items];
			newItemName = '';
		} catch { }
	}

	async function toggleItem(item: AppListItem) {
		try {
			await api.patch(`/api/v1/households/${familyID}/lists/${activeListID}/items/${item.id}`, {
				name: item.name, checked: !item.checked,
			});
			items = items.map((i) => (i.id === item.id ? { ...i, checked: !item.checked } : i));
		} catch { }
	}

	async function deleteItem(itemID: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/lists/${activeListID}/items/${itemID}`);
			items = items.filter((i) => i.id !== itemID);
		} catch { }
	}

	async function clearChecked() {
		try {
			await api.delete(`/api/v1/households/${familyID}/lists/${activeListID}/items/checked`);
			items = items.filter((i) => !i.checked);
		} catch { }
	}

	async function submitRenameList() {
		if (!renamingListID || !renameListValue.trim()) { renamingListID = null; return; }
		try {
			await api.patch(`/api/v1/households/${familyID}/lists/${renamingListID}`, { name: renameListValue.trim() });
			lists = lists.map(l => l.id === renamingListID ? { ...l, name: renameListValue.trim() } : l);
		} catch { }
		renamingListID = null;
	}

	async function submitRenameItem() {
		if (!renamingItemID || !renameItemValue.trim()) { renamingItemID = null; return; }
		const item = items.find(i => i.id === renamingItemID);
		if (!item) { renamingItemID = null; return; }
		try {
			await api.patch(`/api/v1/households/${familyID}/lists/${activeListID}/items/${renamingItemID}`, { name: renameItemValue.trim(), checked: item.checked });
			items = items.map(i => i.id === renamingItemID ? { ...i, name: renameItemValue.trim() } : i);
		} catch { }
		renamingItemID = null;
	}

	const uncheckedItems = $derived(
		items
			.filter((i) => !i.checked)
			.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()),
	);

	const checkedItems = $derived(
		items
			.filter((i) => i.checked)
			.sort((a, b) => new Date(b.checked_at ?? b.created_at).getTime() - new Date(a.checked_at ?? a.created_at).getTime()),
	);
</script>

<!-- List tabs -->
<div class="sticky top-0 z-10 bg-background px-4 md:px-6 pt-4 md:pt-6 pb-2">
<div class="flex items-center gap-2 overflow-x-auto py-1 -my-1 pb-2">
	<div class="flex items-center gap-1.5 flex-1">
		{#each lists as list (list.id)}
			<div class="shrink-0">
				{#if activeListID === list.id && lists.length > 1}
					<div class="flex items-center gap-1 pl-3 pr-1.5 py-1.5 rounded-full bg-primary text-primary-foreground">
						{#if renamingListID === list.id}
							<input
								class="text-sm font-medium bg-transparent border-none outline-none w-28 text-primary-foreground placeholder:text-primary-foreground/60"
								bind:value={renameListValue}
								onblur={submitRenameList}
								onkeydown={(e) => { if (e.key === 'Enter') submitRenameList(); if (e.key === 'Escape') renamingListID = null; }}
								use:focusSelect
							/>
						{:else}
							<button
								onclick={() => (activeListID = list.id)}
								ondblclick={() => { renamingListID = list.id; renameListValue = list.name; }}
								class="text-sm font-medium whitespace-nowrap cursor-pointer"
							>{list.name}</button>
						{/if}
						<button
							onclick={() => (confirmDeleteList = list)}
							class="flex items-center justify-center w-5 h-5 rounded-full hover:bg-primary-foreground/20 cursor-pointer"
							aria-label="Delete list"
						>
							<X class="w-3 h-3" />
						</button>
					</div>
				{:else}
					<button
						onclick={() => (activeListID = list.id)}
						class="px-3 py-1.5 rounded-full text-sm font-medium transition-colors cursor-pointer whitespace-nowrap
							{activeListID === list.id ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'}"
					>{list.name}</button>
				{/if}
			</div>
		{/each}
	</div>

	{#if addingList}
		<form onsubmit={createList} class="flex items-center gap-1.5 shrink-0">
			<Input
				bind:value={newListName}
				placeholder="List name…"
				class="h-8 text-sm w-32"
				onkeydown={(e) => { if (e.key === 'Escape') addingList = false; }}
			/>
			<Button size="sm" class="h-8" type="submit" disabled={!newListName.trim()}>Add</Button>
			<Button size="sm" variant="ghost" class="h-8" onclick={() => (addingList = false)}>Cancel</Button>
		</form>
	{:else}
		<Button size="sm" variant="outline" class="shrink-0" onclick={() => (addingList = true)}>+ List</Button>
	{/if}
</div>
</div>

<div class="px-4 md:px-6 pb-8 pt-3">
<!-- Delete list confirmation -->
{#if confirmDeleteList}
	<div class="mb-4 p-3 rounded-lg border border-destructive/30 bg-destructive/5 flex items-center justify-between gap-3 text-sm">
		<span>Delete <strong>{confirmDeleteList.name}</strong>? This will remove all {items.length} item{items.length === 1 ? '' : 's'}.</span>
		<div class="flex gap-2 shrink-0">
			<Button size="sm" variant="destructive" onclick={() => deleteList(confirmDeleteList!.id)}>Delete</Button>
			<Button size="sm" variant="ghost" onclick={() => (confirmDeleteList = null)}>Cancel</Button>
		</div>
	</div>
{/if}

{#if lists.length === 0}
	<div class="flex flex-col items-center gap-2 py-16 text-muted-foreground">
		<ShoppingCart class="w-10 h-10 opacity-30" />
		<p class="text-sm font-medium">No lists yet</p>
		<p class="text-xs">Create a Shopping list or any other list above.</p>
	</div>
{:else if activeListID}
	<!-- Add item -->
	<form onsubmit={addItem} class="mb-4">
		<Input bind:value={newItemName} placeholder="Add item… (press Enter)" class="bg-muted/20 border-dashed focus-visible:border-solid" />
	</form>

	{#if uncheckedItems.length === 0 && checkedItems.length === 0}
		<div class="flex flex-col items-center gap-2 py-12 text-muted-foreground">
			<ShoppingCart class="w-8 h-8 opacity-30" />
			<p class="text-sm font-medium">List is empty</p>
			<p class="text-xs">Add items above to get started.</p>
		</div>
	{:else}
		<!-- To buy -->
		{#if uncheckedItems.length > 0}
			<div class="flex items-center gap-3 mb-2">
				<span class="text-xs font-semibold uppercase tracking-wide text-muted-foreground shrink-0">To buy ({uncheckedItems.length})</span>
				<div class="flex-1 h-px bg-border"></div>
			</div>
			<div class="flex flex-col divide-y divide-border mb-5">
				{#each uncheckedItems as item (item.id)}
					<div class="flex items-center gap-3 py-3">
						<Checkbox checked={false} onCheckedChange={() => toggleItem(item)} />
						{#if renamingItemID === item.id}
							<input
								class="flex-1 text-sm bg-transparent border-none outline-none border-b border-primary"
								bind:value={renameItemValue}
								onblur={submitRenameItem}
								onkeydown={(e) => { if (e.key === 'Enter') submitRenameItem(); if (e.key === 'Escape') renamingItemID = null; }}
								use:focusSelect
							/>
						{:else}
							<span
								class="flex-1 text-sm cursor-text"
								ondblclick={() => { renamingItemID = item.id; renameItemValue = item.name; }}
							>{item.name}</span>
						{/if}
						<button onclick={() => deleteItem(item.id)} class="text-muted-foreground hover:text-destructive transition-colors cursor-pointer p-1" aria-label="Delete item">
							<X class="w-3.5 h-3.5" />
						</button>
					</div>
				{/each}
			</div>
		{:else if checkedItems.length > 0}
			<div class="py-8 text-center text-sm text-muted-foreground">
				All done! Tap "Clear all" when you're home.
			</div>
		{/if}

		<!-- In cart -->
		{#if checkedItems.length > 0}
			<div class="flex items-center justify-between mb-2">
				<div class="flex items-center gap-3 flex-1">
					<span class="text-xs font-semibold uppercase tracking-wide text-muted-foreground shrink-0">In cart ({checkedItems.length})</span>
					<div class="flex-1 h-px bg-border"></div>
				</div>
				<button onclick={clearChecked} class="text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer ml-3">
					Clear all
				</button>
			</div>
			<div class="flex flex-col divide-y divide-border opacity-50">
				{#each checkedItems as item (item.id)}
					<div class="flex items-center gap-3 py-3">
						<Checkbox checked={true} onCheckedChange={() => toggleItem(item)} />
						<span class="flex-1 text-sm line-through">{item.name}</span>
						<button onclick={() => deleteItem(item.id)} class="text-muted-foreground hover:text-destructive transition-colors cursor-pointer p-1" aria-label="Delete item">
							<X class="w-3.5 h-3.5" />
						</button>
					</div>
				{/each}
			</div>
		{/if}
	{/if}
{/if}
</div>
