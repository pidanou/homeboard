<script lang="ts">
	import type { Task, Member, AppLabel } from '$lib/types';
	import { chipClass, dotClass } from '$lib/labels';
	import { relativeDate } from '$lib/dates';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { User } from 'lucide-svelte';

	let { task, members, labels, isDoneFilter, onclick, ontoggle }: {
		task: Task;
		members: Member[];
		labels: AppLabel[];
		isDoneFilter: boolean;
		onclick: () => void;
		ontoggle: (e: MouseEvent) => void;
	} = $props();

	const priorityBorder: Record<string, string> = {
		high: 'border-l-4 border-l-red-500',
		medium: 'border-l-4 border-l-yellow-400',
	};

	function memberName(uid: string | undefined): string | null {
		if (!uid) return null;
		return members.find((m) => m.user_id === uid)?.name ?? null;
	}

	function labelByID(id: string): AppLabel | undefined {
		return labels.find((l) => l.id === id);
	}

	function isOverdue(t: Task): boolean {
		return !!(t.end_date && t.status !== 'done' && new Date(t.end_date) < new Date());
	}
</script>

<button
	class="w-full text-left flex items-start gap-3 rounded-lg border border-border bg-card px-4 py-3 shadow-sm hover:bg-accent/50 transition-colors cursor-pointer
		{isDoneFilter
			? 'border-l-4 border-l-slate-300 dark:border-l-slate-600 opacity-60'
			: (priorityBorder[task.priority] ?? '')}"
	{onclick}
>
	<div role="presentation" onclick={ontoggle} class="mt-0.5">
		<Checkbox checked={task.status === 'done'} class="pointer-events-none" />
	</div>
	<div class="flex-1 min-w-0">
		<p class="text-sm font-medium truncate {task.status === 'done' ? 'line-through text-muted-foreground' : ''}">
			{task.title}
		</p>
		{#if task.description && task.status !== 'done'}
			<p class="text-xs text-muted-foreground truncate mt-0.5">{task.description}</p>
		{/if}
		{#if task.status !== 'done'}
			<div class="flex items-center gap-2 mt-1 flex-wrap">
				{#if task.end_date}
					<p class="text-xs {isOverdue(task) ? 'text-destructive font-medium' : 'text-muted-foreground'}">
						Due {relativeDate(task.end_date)}
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
				{#each task.label_ids ?? [] as lid}
					{@const lbl = labelByID(lid)}
					{#if lbl}
						<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium {chipClass(lbl.color)}">
							<span class="w-1.5 h-1.5 rounded-full {dotClass(lbl.color)} shrink-0"></span>
							{lbl.name}
						</span>
					{/if}
				{/each}
			</div>
		{/if}
	</div>
</button>
