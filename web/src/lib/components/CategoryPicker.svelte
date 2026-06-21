<script lang="ts">
	import type { AppCategory } from '$lib/types';
	import { chipClass, dotClass, LABEL_COLORS, type LabelColor } from '$lib/categories';
	import { api } from '$lib/api/client';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';

	let { familyID, categories, selectedID = $bindable<string | undefined>(undefined) }: {
		familyID: string;
		categories: AppCategory[];
		selectedID: string | undefined;
	} = $props();

	let created = $state<AppCategory[]>([]);
	const allCategories = $derived([...categories, ...created.filter((c) => !categories.some((l) => l.id === c.id))]);

	let adding = $state(false);
	let newName = $state('');
	let newColor = $state<LabelColor>('blue');

	function select(id: string) {
		selectedID = selectedID === id ? undefined : id;
	}

	async function createCategory() {
		if (!newName.trim()) return;
		try {
			const cat = await api.post<AppCategory>(`/api/v1/households/${familyID}/categories`, {
				name: newName.trim(),
				color: newColor,
			});
			created = [...created, cat];
			selectedID = cat.id;
			newName = '';
			newColor = 'blue';
			adding = false;
		} catch { }
	}
</script>

<div class="flex flex-col gap-2">
	{#if allCategories.length > 0}
		<div class="flex flex-wrap gap-1.5">
			{#each allCategories as cat}
				<button
					type="button"
					onclick={() => select(cat.id)}
					class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-all cursor-pointer
						{selectedID === cat.id ? chipClass(cat.color) : 'opacity-40 ' + chipClass(cat.color)}"
				>
					<span class="w-1.5 h-1.5 rounded-full {dotClass(cat.color)} shrink-0"></span>
					{cat.name}
				</button>
			{/each}
		</div>
	{/if}

	{#if adding}
		<div class="flex flex-col gap-2 p-2.5 rounded-lg border border-border bg-muted/30">
			<Input
				bind:value={newName}
				placeholder="Category name…"
				class="h-7 text-xs"
				onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); createCategory(); } if (e.key === 'Escape') adding = false; }}
			/>
			<div class="flex items-center gap-1.5">
				{#each LABEL_COLORS as color}
					<button
						type="button"
						aria-label={color}
						onclick={() => (newColor = color)}
						class="w-4 h-4 rounded-full transition-transform cursor-pointer {dotClass(color)}
							{newColor === color ? 'ring-2 ring-offset-1 ring-foreground scale-110' : ''}"
					></button>
				{/each}
			</div>
			<div class="flex gap-1.5">
				<Button size="sm" class="h-6 text-xs px-2" onclick={createCategory} disabled={!newName.trim()}>Add</Button>
				<Button size="sm" variant="ghost" class="h-6 text-xs px-2" onclick={() => (adding = false)}>Cancel</Button>
			</div>
		</div>
	{:else}
		<button
			type="button"
			onclick={() => (adding = true)}
			class="text-xs text-muted-foreground hover:text-foreground transition-colors text-left cursor-pointer w-fit"
		>+ New category</button>
	{/if}
</div>
