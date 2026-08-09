<script lang="ts">
	import type { AppCategory } from '$lib/types';
	import { chipStyle, dotStyle, resolveHex, LABEL_COLORS } from '$lib/categories';
	import { api } from '$lib/api/client';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import ColorPicker from 'svelte-awesome-color-picker';
	import { Palette } from 'lucide-svelte';
	import * as msg from '$lib/paraglide/messages.js';

	let { familyID, categories, selectedID = $bindable<string | undefined>(undefined) }: {
		familyID: string;
		categories: AppCategory[];
		selectedID: string | undefined;
	} = $props();

	let created = $state<AppCategory[]>([]);
	const allCategories = $derived([...categories, ...created.filter((c) => !categories.some((l) => l.id === c.id))]);

	let adding = $state(false);
	let newName = $state('');
	let newColor = $state('blue');

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
						{selectedID === cat.id ? '' : 'opacity-40'}"
					style={chipStyle(cat.color)}
				>
					<span class="w-1.5 h-1.5 rounded-full shrink-0" style={dotStyle(cat.color)}></span>
					{cat.name}
				</button>
			{/each}
		</div>
	{/if}

	{#if adding}
		<div class="flex flex-col gap-2 p-2.5 rounded-lg border border-border bg-muted/30">
			<Input
				bind:value={newName}
				placeholder={msg.category_picker_name_placeholder()}
				class="h-7 text-xs"
				onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); createCategory(); } if (e.key === 'Escape') adding = false; }}
			/>
			<div class="flex items-center gap-1.5">
				{#each LABEL_COLORS as color}
					<button
						type="button"
						aria-label={color}
						onclick={() => (newColor = color)}
						style={dotStyle(color)}
						class="w-4 h-4 rounded-full transition-transform cursor-pointer
							{newColor === color ? 'ring-2 ring-offset-1 ring-foreground scale-110' : ''}"
					></button>
				{/each}
				<div class="relative" title="Custom color" style="--cp-bg-color: var(--popover); --cp-border-color: var(--border); --cp-text-color: var(--foreground); --cp-input-color: var(--input); --cp-button-hover-color: var(--muted); --focus-color: var(--ring); --input-size: 16px; --picker-z-index: 50;">
					<ColorPicker
						hex={resolveHex(newColor)}
						label=""
						isAlpha={false}
						onInput={(c) => c.hex && (newColor = c.hex)}
					/>
					<div class="absolute inset-0 flex items-center justify-center pointer-events-none">
						<Palette class="w-2.5 h-2.5 text-white drop-shadow-[0_1px_1px_rgba(0,0,0,0.8)]" />
					</div>
				</div>
			</div>
			<div class="flex gap-1.5">
				<Button size="sm" class="h-6 text-xs px-2" onclick={createCategory} disabled={!newName.trim()}>{msg.action_add()}</Button>
				<Button size="sm" variant="ghost" class="h-6 text-xs px-2" onclick={() => (adding = false)}>{msg.dialog_cancel()}</Button>
			</div>
		</div>
	{:else}
		<button
			type="button"
			onclick={() => (adding = true)}
			class="text-xs text-muted-foreground hover:text-foreground transition-colors text-left cursor-pointer w-fit"
		>+ {msg.category_picker_new()}</button>
	{/if}
</div>
