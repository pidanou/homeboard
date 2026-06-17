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
	let error = $state('');
	let newItemName = $state('');
	let addingList = $state(false);
	let newListName = $state('');
	let confirmDeleteList = $state<AppList | null>(null);

	let es: EventSource | null = null;
	let errorTimer: ReturnType<typeof setTimeout> | null = null;

	async function loadLists() {
		const res = await api.get<AppList[]>(`/api/v1/families/${familyID}/lists`);
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
		const res = await api.get<AppListItem[]>(`/api/v1/families/${familyID}/lists/${activeListID}/items`);
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
		es = new EventSource(sseUrl(`/api/v1/families/${familyID}/stream`) as string);
		es.onmessage = (e) => { if (e.data === 'refresh') loadAll(); };
		es.onerror = () => { es?.close(); es = null; };
	});

	onDestroy(() => {
		es?.close();
		if (errorTimer) clearTimeout(errorTimer);
	});

	function setError(err: unknown) {
		error = err instanceof Error ? err.message : 'Something went wrong';
		if (errorTimer) clearTimeout(errorTimer);
		errorTimer = setTimeout(() => (error = ''), 4000);
	}

	async function createList(e: SubmitEvent) {
		e.preventDefault();
		if (!newListName.trim()) return;
		try {
			const list = await api.post<AppList>(`/api/v1/families/${familyID}/lists`, { name: newListName.trim() });
			lists = [...lists, list];
			activeListID = list.id;
			newListName = '';
			addingList = false;
		} catch (err) {
			setError(err);
		}
	}

	async function deleteList(id: string) {
		try {
			await api.delete(`/api/v1/families/${familyID}/lists/${id}`);
			lists = lists.filter((l) => l.id !== id);
			confirmDeleteList = null;
			if (activeListID === id) {
				activeListID = lists[0]?.id ?? '';
				items = [];
			}
		} catch (err) {
			setError(err);
		}
	}

	async function addItem(e: SubmitEvent) {
		e.preventDefault();
		if (!newItemName.trim() || !activeListID) return;
		try {
			const item = await api.post<AppListItem>(`/api/v1/families/${familyID}/lists/${activeListID}/items`, { name: newItemName.trim() });
			items = [item, ...items];
			newItemName = '';
		} catch (err) {
			setError(err);
		}
	}

	async function toggleItem(item: AppListItem) {
		try {
			await api.patch(`/api/v1/families/${familyID}/lists/${activeListID}/items/${item.id}`, {
				name: item.name, checked: !item.checked,
			});
			items = items.map((i) => (i.id === item.id ? { ...i, checked: !item.checked } : i));
		} catch (err) {
			setError(err);
		}
	}

	async function deleteItem(itemID: string) {
		try {
			await api.delete(`/api/v1/families/${familyID}/lists/${activeListID}/items/${itemID}`);
			items = items.filter((i) => i.id !== itemID);
		} catch (err) {
			setError(err);
		}
	}

	async function clearChecked() {
		try {
			await api.delete(`/api/v1/families/${familyID}/lists/${activeListID}/items/checked`);
			items = items.filter((i) => !i.checked);
		} catch (err) {
			setError(err);
		}
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

{#if error}
	<div class="flex items-center justify-between gap-2 px-3 py-2 mb-3 rounded-md bg-destructive/10 text-destructive text-sm">
		<span>{error}</span>
		<button onclick={() => (error = '')} class="shrink-0 opacity-70 hover:opacity-100">✕</button>
	</div>
{/if}

<!-- List tabs -->
<div class="flex items-center gap-2 mb-4 overflow-x-auto pb-1">
	<div class="flex items-center gap-1.5 flex-1">
		{#each lists as list (list.id)}
			<div class="relative group shrink-0">
				<button
					onclick={() => (activeListID = list.id)}
					class="px-3 py-1.5 rounded-full text-sm font-medium transition-colors cursor-pointer whitespace-nowrap
						{activeListID === list.id ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground hover:bg-muted/80'}"
				>{list.name}</button>
				{#if activeListID === list.id && lists.length > 1}
					<button
						onclick={() => (confirmDeleteList = list)}
						class="absolute -top-1 -right-1 w-4 h-4 rounded-full bg-muted-foreground text-background
							flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
						aria-label="Delete list"
					>
						<X class="w-2.5 h-2.5" />
					</button>
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
			<p class="text-sm">List is empty. Add items above.</p>
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
						<span class="flex-1 text-sm">{item.name}</span>
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
