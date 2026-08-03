<script lang="ts">
	import type { Task, CalEvent, Member, AppCategory } from '$lib/types';
	import { localDayMs } from '$lib/dates';
	import TaskCard from '$lib/components/TaskCard.svelte';
	import EventCard from '$lib/components/EventCard.svelte';
	import * as m from '$lib/paraglide/messages.js';

	let { events, tasks, members, categories, todayMs, now, onTaskClick, onEventClick, ontoggle }: {
		events: CalEvent[];
		tasks: Task[];
		members: Member[];
		categories: AppCategory[];
		todayMs: number;
		now: Date;
		onTaskClick: (t: Task) => void;
		onEventClick: (e: CalEvent) => void;
		ontoggle: (t: Task, e: MouseEvent) => void;
	} = $props();

	const weekEndMs = $derived(todayMs + 7 * 86400000);

	type Item = { kind: 'event'; date: number; data: CalEvent } | { kind: 'task'; date: number; data: Task };

	const items = $derived<Item[]>(
		[
			...events
				.filter((e) => localDayMs(e.start_at) > todayMs && localDayMs(e.start_at) <= weekEndMs)
				.map((data): Item => ({ kind: 'event', date: localDayMs(data.start_at), data })),
			...tasks
				.filter((t) => t.status !== 'done' && t.end_date && localDayMs(t.end_date) > todayMs && localDayMs(t.end_date) <= weekEndMs)
				.map((data): Item => ({ kind: 'task', date: localDayMs(data.end_date!), data })),
		].sort((a, b) => a.date - b.date),
	);
</script>

<div>
	<div class="flex items-center gap-3 mb-3">
		<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-foreground">{m.today_upcoming_label()}</span>
		<div class="flex-1 h-px bg-border"></div>
	</div>
	{#if items.length > 0}
		<div class="flex flex-col gap-2">
			{#each items as item (item.kind + item.data.id)}
				{#if item.kind === 'event'}
					<EventCard event={item.data} {members} {categories} {now} onclick={() => onEventClick(item.data)} />
				{:else}
					<TaskCard
						task={item.data}
						{members}
						{categories}
						isDoneFilter={false}
						onclick={() => onTaskClick(item.data)}
						ontoggle={(e) => ontoggle(item.data, e)}
					/>
				{/if}
			{/each}
		</div>
	{:else}
		<p class="text-sm text-muted-foreground/50 italic">{m.today_upcoming_empty()}</p>
	{/if}
</div>
