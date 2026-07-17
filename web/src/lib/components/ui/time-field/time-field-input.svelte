<script lang="ts">
	import { TimeField as TimeFieldPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		class: className,
		...restProps
	}: TimeFieldPrimitive.InputProps = $props();
</script>

<TimeFieldPrimitive.Input
	class={cn(
		'bg-input/30 border-input focus-within:border-ring focus-within:ring-ring/50 flex h-9 w-full items-center rounded-4xl border px-3 py-1 text-base transition-colors focus-within:ring-[3px] md:text-sm',
		className
	)}
	{...restProps}
>
	{#snippet children({ segments })}
		{#each segments as { part, value }, i (i)}
			<TimeFieldPrimitive.Segment
				{part}
				class={cn(
					'focus:bg-accent focus:text-accent-foreground rounded-sm px-0.5 tabular-nums outline-none',
					part === 'literal' && 'text-muted-foreground px-0.5'
				)}
			>
				{value}
			</TimeFieldPrimitive.Segment>
		{/each}
	{/snippet}
</TimeFieldPrimitive.Input>
