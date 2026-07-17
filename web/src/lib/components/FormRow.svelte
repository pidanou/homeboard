<script lang="ts">
	import type { ComponentType, SvelteComponent, Snippet } from 'svelte';
	import { Checkbox } from '$lib/components/ui/checkbox';

	let { icon: Icon, checked = $bindable(), compact = false, children }: {
		icon?: ComponentType<SvelteComponent<{ class?: string }>>;
		checked?: boolean;
		compact?: boolean;
		children: Snippet;
	} = $props();

	const isCheckbox = $derived(checked !== undefined);
</script>

{#if isCheckbox}
	<label class="flex items-start gap-3 cursor-pointer">
		<div class="{compact ? 'h-5' : 'h-9'} w-4 flex items-center justify-center shrink-0">
			<Checkbox bind:checked />
		</div>
		<div class="{compact ? 'h-5' : 'h-9'} flex-1 min-w-0 flex items-center text-sm">
			{@render children()}
		</div>
	</label>
{:else}
	<div class="flex items-start gap-3">
		<div class="{compact ? 'h-5' : 'h-9'} w-4 flex items-center justify-center shrink-0">
			{#if Icon}
				<Icon class="w-4 h-4 text-muted-foreground" />
			{/if}
		</div>
		<div class="flex-1 min-w-0">
			{@render children()}
		</div>
	</div>
{/if}
