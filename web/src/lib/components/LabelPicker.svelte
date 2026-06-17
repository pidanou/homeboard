<script lang="ts">
	import type { AppLabel } from '$lib/types';
	import { chipClass, dotClass, LABEL_COLORS, type LabelColor } from '$lib/labels';
	import { api } from '$lib/api/client';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';

	let { familyID, labels, selectedIDs = $bindable([]), onError }: {
		familyID: string;
		labels: AppLabel[];
		selectedIDs: string[];
		onError: (e: unknown) => void;
	} = $props();

	// locally created labels, merged with prop so page refresh isn't required
	let created = $state<AppLabel[]>([]);
	const allLabels = $derived([...labels, ...created.filter((c) => !labels.some((l) => l.id === c.id))]);

	let adding = $state(false);
	let newName = $state('');
	let newColor = $state<LabelColor>('blue');

	function toggle(id: string) {
		selectedIDs = selectedIDs.includes(id)
			? selectedIDs.filter((x) => x !== id)
			: [...selectedIDs, id];
	}

	async function createLabel() {
		if (!newName.trim()) return;
		try {
			const lbl = await api.post<AppLabel>(`/api/v1/families/${familyID}/labels`, {
				name: newName.trim(),
				color: newColor,
			});
			created = [...created, lbl];
			selectedIDs = [...selectedIDs, lbl.id];
			newName = '';
			newColor = 'blue';
			adding = false;
		} catch (e) {
			onError(e);
		}
	}
</script>

<div class="flex flex-col gap-2">
	{#if allLabels.length > 0}
		<div class="flex flex-wrap gap-1.5">
			{#each allLabels as lbl}
				<button
					type="button"
					onclick={() => toggle(lbl.id)}
					class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-all cursor-pointer
						{selectedIDs.includes(lbl.id) ? chipClass(lbl.color) : 'opacity-40 ' + chipClass(lbl.color)}"
				>
					<span class="w-1.5 h-1.5 rounded-full {dotClass(lbl.color)} shrink-0"></span>
					{lbl.name}
				</button>
			{/each}
		</div>
	{/if}

	{#if adding}
		<div class="flex flex-col gap-2 p-2.5 rounded-lg border border-border bg-muted/30">
			<Input
				bind:value={newName}
				placeholder="Label name…"
				class="h-7 text-xs"
				onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); createLabel(); } if (e.key === 'Escape') adding = false; }}
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
				<Button size="sm" class="h-6 text-xs px-2" onclick={createLabel} disabled={!newName.trim()}>Add</Button>
				<Button size="sm" variant="ghost" class="h-6 text-xs px-2" onclick={() => (adding = false)}>Cancel</Button>
			</div>
		</div>
	{:else}
		<button
			type="button"
			onclick={() => (adding = true)}
			class="text-xs text-muted-foreground hover:text-foreground transition-colors text-left cursor-pointer w-fit"
		>+ New label</button>
	{/if}
</div>
