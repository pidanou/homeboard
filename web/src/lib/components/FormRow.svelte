<script lang="ts">
	import type { ComponentType, SvelteComponent, Snippet } from 'svelte';

	let { icon: Icon, align = 'center', children }: {
		icon: ComponentType<SvelteComponent<{ class?: string }>>;
		align?: 'center' | 'start';
		children: Snippet;
	} = $props();
</script>

<div class="flex {align === 'start' ? 'items-start' : 'items-center'} gap-3">
	<Icon class="w-4 h-4 text-muted-foreground shrink-0 {align === 'start' ? 'mt-2.5' : ''}" />
	<div class="flex-1 min-w-0">
		{@render children()}
	</div>
</div>
