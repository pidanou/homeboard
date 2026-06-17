<script lang="ts">
	import type { CalEvent, Member, AppLabel } from '$lib/types';
	import { chipClass, dotClass } from '$lib/labels';
	import { fmtDate, fmtDateTime } from '$lib/dates';
	import { AlarmClock, MapPin, User } from 'lucide-svelte';

	let { event, members, labels, now, onclick }: {
		event: CalEvent;
		members: Member[];
		labels: AppLabel[];
		now: Date;
		onclick: () => void;
	} = $props();

	function memberName(uid: string): string | null {
		return members.find((m) => m.user_id === uid)?.name ?? null;
	}

	function labelByID(id: string): AppLabel | undefined {
		return labels.find((l) => l.id === id);
	}
</script>

<button
	class="w-full text-left flex items-start gap-3 rounded-lg border border-border bg-card px-4 py-3 shadow-sm hover:bg-accent/50 transition-colors cursor-pointer"
	{onclick}
>
	<AlarmClock class="w-4 h-4 mt-0.5 shrink-0 text-blue-400" />
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
		{#if event.attendee_ids && event.attendee_ids.length > 0}
			<div class="flex items-center gap-2 mt-1 flex-wrap">
				{#each event.attendee_ids as uid}
					{@const name = memberName(uid)}
					{#if name}
						<span class="inline-flex items-center gap-0.5 text-xs text-muted-foreground">
							<User class="w-3 h-3" />{name}
						</span>
					{/if}
				{/each}
			</div>
		{/if}
		{#if event.label_ids && event.label_ids.length > 0}
			<div class="flex items-center gap-1.5 mt-1 flex-wrap">
				{#each event.label_ids as lid}
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
