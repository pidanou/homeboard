<script lang="ts">
	import type { ComponentType, SvelteComponent, Snippet } from 'svelte';

	let { icon: Icon, children }: {
		icon: ComponentType<SvelteComponent<{ class?: string }>>;
		children: Snippet;
	} = $props();
</script>

<div class="flex items-start gap-3">
	<div class="h-9 flex items-center justify-center shrink-0">
		<Icon class="w-4 h-4 text-muted-foreground" />
	</div>
	<div class="flex-1 min-w-0">
		{@render children()}
	</div>
</div>
