<script lang="ts">
	import type { Task, Member, AppCategory } from '$lib/types';
	import { chipStyle, dotStyle } from '$lib/categories';
	import { relativeDate, taskHasTime, fmtTime, localDayMs } from '$lib/dates';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { User, Star } from 'lucide-svelte';
	import * as msg from '$lib/paraglide/messages.js';

	let { task, members, categories, isDoneFilter, onclick, ontoggle, interactive = true }: {
		task: Task;
		members: Member[];
		categories: AppCategory[];
		isDoneFilter: boolean;
		onclick?: () => void;
		ontoggle: (e: MouseEvent) => void;
		interactive?: boolean;
	} = $props();

	function memberName(uid: string | undefined): string | null {
		if (!uid) return null;
		return members.find((m) => m.user_id === uid)?.name ?? null;
	}

	const category = $derived(categories.find((c) => c.id === task.category_id));

	function isOverdue(t: Task): boolean {
		const today = new Date(); today.setHours(0, 0, 0, 0);
		return !!(t.end_date && t.status !== 'done' && localDayMs(t.end_date) < today.getTime());
	}
</script>

<svelte:element
	this={interactive ? 'button' : 'div'}
	class="w-full text-left flex items-start gap-3 transition-colors {isDoneFilter ? 'opacity-60' : ''}
		{interactive ? 'rounded-lg border border-border bg-card px-4 py-3 hover:bg-accent/50 cursor-pointer' : ''}"
	onclick={interactive ? onclick : undefined}
	role={interactive ? 'button' : undefined}
	tabindex={interactive ? 0 : undefined}
>
	<div role="presentation" onclick={ontoggle} class="mt-0.5">
		<Checkbox checked={task.status === 'done'} class="pointer-events-none" />
	</div>
	<div class="flex-1 min-w-0">
		<p class="text-sm font-medium truncate {task.status === 'done' ? 'line-through text-muted-foreground' : ''}">
			{#if task.important && task.status !== 'done'}
				<Star class="inline w-3 h-3 fill-amber-400 text-amber-400 mr-1 -mt-0.5" />
			{/if}{task.title}
		</p>
		{#if task.description && task.status !== 'done'}
			<p class="text-xs text-muted-foreground truncate mt-0.5">{task.description}</p>
		{/if}
		{#if task.status !== 'done'}
			<div class="flex items-center gap-2 mt-1 flex-wrap">
				{#if task.end_date}
					<p class="text-xs {isOverdue(task) ? 'text-destructive font-medium' : 'text-muted-foreground'}">
						{taskHasTime(task.end_date)
							? msg.taskcard_due_at({ date: relativeDate(task.end_date), time: fmtTime(task.end_date) })
							: msg.taskcard_due({ date: relativeDate(task.end_date) })}
					</p>
				{/if}
				{#if task.assigned_to}
					{@const name = memberName(task.assigned_to)}
					{#if name}
						<span class="inline-flex items-center gap-1 text-xs text-muted-foreground">
							<User class="w-3 h-3" />{name}
						</span>
					{/if}
				{/if}
				{#if category}
					<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium" style={chipStyle(category.color)}>
						<span class="w-1.5 h-1.5 rounded-full shrink-0" style={dotStyle(category.color)}></span>
						{category.name}
					</span>
				{/if}
			</div>
		{/if}
	</div>
</svelte:element>
