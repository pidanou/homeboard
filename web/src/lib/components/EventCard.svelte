<script lang="ts">
	import type { CalEvent, Member, AppCategory } from '$lib/types';
	import { chipClass, dotClass } from '$lib/categories';
	import { fmtDate, fmtDateTime } from '$lib/dates';
	import { MapPin, User, CalendarDays } from 'lucide-svelte';

	let { event, members, categories, now, onclick }: {
		event: CalEvent;
		members: Member[];
		categories: AppCategory[];
		now: Date;
		onclick: () => void;
	} = $props();

	function memberName(uid: string): string | null {
		return members.find((m) => m.user_id === uid)?.name ?? null;
	}

	const category = $derived(categories.find((c) => c.id === event.category_id));
</script>

<button
	class="w-full text-left flex items-start gap-3 rounded-lg border border-border bg-card px-4 py-3 hover:bg-accent/50 transition-colors cursor-pointer"
	{onclick}
>
	<CalendarDays class="w-4 h-4 mt-0.5 shrink-0 text-muted-foreground" />
	<div class="flex-1 min-w-0">
		<div class="flex items-start justify-between gap-2">
			<p class="text-sm font-medium truncate">{event.title}</p>
			{#if new Date(event.start_at) <= now}
				<span class="relative flex h-2 w-2 mt-1.5 shrink-0">
					<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-orange-400 opacity-75"></span>
					<span class="relative inline-flex rounded-full h-2 w-2 bg-orange-500"></span>
				</span>
			{/if}
		</div>
		{#if event.description}
			<p class="text-xs text-muted-foreground truncate mt-0.5">{event.description}</p>
		{/if}
		<p class="text-xs text-muted-foreground mt-1">
			{event.all_day ? fmtDate(event.start_at) : fmtDateTime(event.start_at)}
			→
			{event.all_day ? fmtDate(event.end_at) : fmtDateTime(event.end_at)}
			{#if event.location}
				· <MapPin class="w-3 h-3 inline -mt-px" />{event.location}
			{/if}
		</p>
		<div class="flex items-center gap-2 mt-1 flex-wrap">
			{#if event.attendee_ids && event.attendee_ids.length > 0}
				{#each event.attendee_ids as uid}
					{@const name = memberName(uid)}
					{#if name}
						<span class="inline-flex items-center gap-0.5 text-xs text-muted-foreground">
							<User class="w-3 h-3" />{name}
						</span>
					{/if}
				{/each}
			{/if}
			{#if category}
				<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium {chipClass(category.color)}">
					<span class="w-1.5 h-1.5 rounded-full {dotClass(category.color)} shrink-0"></span>
					{category.name}
				</span>
			{/if}
		</div>
	</div>
</button>
