<script lang="ts">
	import type { ComponentType, SvelteComponent, Snippet } from 'svelte';

	let { icon: Icon, compact = false, children }: {
		icon: ComponentType<SvelteComponent<{ class?: string }>>;
		compact?: boolean;
		children: Snippet;
	} = $props();
</script>

<div class="flex items-start gap-3">
	<div class="{compact ? 'h-5' : 'h-9'} flex items-center justify-center shrink-0">
		<Icon class="w-4 h-4 text-muted-foreground" />
	</div>
	<div class="flex-1 min-w-0">
		{@render children()}
	</div>
</div>
