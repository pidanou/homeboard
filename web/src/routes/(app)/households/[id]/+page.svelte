<script lang="ts">
	import { page } from '$app/stores';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { GripVertical, Plus, Pencil, Trash2 } from 'lucide-svelte';
	import type { Task, CalEvent, Member, AppCategory } from '$lib/types';
	import { localDayMs, fmtTime, fmtWeekdayDate } from '$lib/dates';
	import { sortable } from '$lib/sortable';
	import TaskCard from '$lib/components/TaskCard.svelte';
	import EventCard from '$lib/components/EventCard.svelte';
	import DayView from '$lib/components/DayView.svelte';
	import CreateDialog from '$lib/components/CreateDialog.svelte';
	import EditDialog from '$lib/components/EditDialog.svelte';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import TodayUpcomingCard from '$lib/components/TodayUpcomingCard.svelte';
	import TodayListsCard from '$lib/components/TodayListsCard.svelte';
	import * as Popover from '$lib/components/ui/popover';
	import * as m from '$lib/paraglide/messages.js';

	const familyID = $derived($page.params.id ?? '');
	const now = new Date();
	const todayMs = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
	const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
	const todayEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999);
	const weekEnd = new Date(todayEnd.getTime() + 7 * 86400000);

	let members = $state<Member[]>([]);
	let tasks = $state<Task[]>([]);
	let events = $state<CalEvent[]>([]);
	let categories = $state<AppCategory[]>([]);

	let createDialog: { open: (t?: 'task' | 'event') => void } | undefined = $state();
	let editDialog: {
		openTask: (t: Task) => void; openEvent: (e: CalEvent) => void;
		deleteTask: (t: Task) => void; deleteEvent: (e: CalEvent) => void;
	} | undefined = $state();

	let previewOpen = $state(false);
	let previewAnchor = $state<{ getBoundingClientRect: () => DOMRect } | null>(null);
	let previewItem = $state<{ kind: 'task'; data: Task } | { kind: 'event'; data: CalEvent } | null>(null);

	function showPreview(kind: 'task' | 'event', data: Task | CalEvent, jsEvent: MouseEvent) {
		const { clientX, clientY } = jsEvent;
		previewAnchor = { getBoundingClientRect: () => new DOMRect(clientX, clientY, 0, 0) };
		previewItem = { kind, data } as typeof previewItem;
		previewOpen = true;
	}

	let es: EventSource | null = null;

	async function loadData() {
		const [membersRes, tasksRes, eventsRes, catsRes] = await Promise.allSettled([
			api.get<Member[]>(`/api/v1/households/${familyID}/members`),
			api.get<Task[]>(`/api/v1/households/${familyID}/tasks`),
			api.get<CalEvent[]>(`/api/v1/households/${familyID}/events?from=${todayStart.toISOString()}&to=${weekEnd.toISOString()}`),
			api.get<AppCategory[]>(`/api/v1/households/${familyID}/categories`),
		]);
		if (membersRes.status === 'fulfilled') members = membersRes.value ?? [];
		if (tasksRes.status === 'fulfilled') tasks = tasksRes.value ?? [];
		if (eventsRes.status === 'fulfilled') events = eventsRes.value ?? [];
		if (catsRes.status === 'fulfilled') categories = catsRes.value ?? [];
	}

	$effect(() => {
		loadData();
		es?.close();
		es = new EventSource(sseUrl(`/api/v1/households/${familyID}/stream`));
		es.onmessage = (e) => { if (e.data === 'refresh') loadData(); };
		es.onerror = () => { es?.close(); es = null; };
		return () => es?.close();
	});

	async function toggleTask(task: Task, e: MouseEvent) {
		e.stopPropagation();
		const newStatus = task.status === 'done' ? 'todo' : 'done';
		try {
			await api.patch(`/api/v1/households/${familyID}/tasks/${task.id}`, {
				title: task.title, description: task.description, important: task.important,
				status: newStatus, assigned_to: task.assigned_to, end_date: task.end_date, category_id: task.category_id,
			});
			tasks = tasks.map((t) => t.id === task.id ? { ...t, status: newStatus } : t);
		} catch { }
	}

	const sortedEvents = $derived(
		events
			.filter((e) => localDayMs(e.start_at) === todayMs)
			.sort((a, b) => {
				if (a.all_day !== b.all_day) return a.all_day ? -1 : 1;
				return new Date(a.start_at).getTime() - new Date(b.start_at).getTime();
			}),
	);

	let dueTodayTasks = $state<Task[]>([]);
	$effect(() => {
		const overdue = tasks
			.filter((t) => t.status !== 'done' && t.end_date && localDayMs(t.end_date) < todayMs)
			.sort((a, b) => new Date(a.end_date!).getTime() - new Date(b.end_date!).getTime());
		const today = tasks.filter((t) => t.status !== 'done' && t.end_date && localDayMs(t.end_date) === todayMs);
		dueTodayTasks = [...overdue, ...today];
	});

	async function handleEventDrop(ev: CalEvent, start: Date, end: Date, allDay: boolean, revert: () => void) {
		try {
			await api.patch(`/api/v1/households/${familyID}/events/${ev.id}`, {
				title: ev.title, description: ev.description ?? '', location: ev.location ?? '',
				start_at: start.toISOString(), end_at: end.toISOString(),
				all_day: allDay, attendee_ids: ev.attendee_ids ?? [], category_id: ev.category_id,
			});
		} catch { revert(); }
	}

	async function handleTaskDrop(task: Task, newDate: Date, revert: () => void) {
		try {
			await api.patch(`/api/v1/households/${familyID}/tasks/${task.id}`, {
				title: task.title, description: task.description, important: task.important,
				status: task.status, assigned_to: task.assigned_to,
				end_date: newDate.toISOString(), category_id: task.category_id,
			});
		} catch { revert(); }
	}

	async function reorderTodayTasks(ids: string[]) {
		const prev = [...dueTodayTasks];
		dueTodayTasks = ids.map((id) => dueTodayTasks.find((t) => t.id === id)!).filter(Boolean);
		try {
			await api.put(`/api/v1/households/${familyID}/tasks/reorder`, { ids });
		} catch { dueTodayTasks = prev; }
	}
</script>

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="px-4 md:px-6 pt-4 md:pt-6 pb-3 shrink-0 flex items-start justify-between gap-3">
		<div>
			<h1 class="text-xl font-semibold">{m.nav_today()}</h1>
			<p class="text-xs text-muted-foreground">
				{fmtWeekdayDate(now)}
			</p>
		</div>
		{#if members.length > 0}
			<div class="flex items-center -space-x-2 shrink-0 pt-0.5">
				{#each members as member (member.user_id)}
					<UserAvatar name={member.name} avatarUrl={member.avatar_url} userId={member.user_id} size={28} class="ring-2 ring-background" />
				{/each}
			</div>
		{/if}
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto md:overflow-hidden px-4 md:px-6 pb-4">
		<div class="flex flex-col gap-6 md:grid md:grid-cols-2 md:gap-8 md:h-full">

			<!-- Tasks column: scrolls on desktop if needed -->
			<div class="flex flex-col gap-6 md:overflow-y-auto md:pr-1">
				<!-- Due today -->
				<div>
					<div class="flex items-center gap-3 mb-3">
						<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-foreground">{m.today_todo_label()}</span>
						<div class="flex-1 h-px bg-border"></div>
						<Button size="sm" variant="ghost" class="h-6 px-2 text-xs text-muted-foreground" onclick={() => createDialog?.open('task', new Date())}>
							<Plus class="w-3 h-3 mr-0.5" />{m.today_add_task_button()}
						</Button>
					</div>
					{#if dueTodayTasks.length > 0}
						<div class="flex flex-col gap-2" use:sortable={{ onReorder: reorderTodayTasks }}>
							{#each dueTodayTasks as task (task.id)}
								<div class="flex items-center gap-1" data-id={task.id}>
									<div data-drag-handle class="cursor-grab active:cursor-grabbing touch-none p-1 shrink-0">
										<GripVertical class="w-4 h-4 text-muted-foreground/40" />
									</div>
									<div class="flex-1 min-w-0">
										<TaskCard {task} {members} {categories} isDoneFilter={false}
											onclick={() => editDialog?.openTask(task)}
											ontoggle={(e) => toggleTask(task, e)} />
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-muted-foreground/50 italic">{m.today_nothing_todo()}</p>
					{/if}
				</div>

				<TodayListsCard {familyID} />

				<TodayUpcomingCard
					{events}
					{tasks}
					{members}
					{categories}
					{todayMs}
					{now}
					onTaskClick={(task) => editDialog?.openTask(task)}
					onEventClick={(event) => editDialog?.openEvent(event)}
					ontoggle={toggleTask}
				/>
			</div>

			<!-- Schedule column: fills remaining height -->
			<div class="flex flex-col md:h-full md:min-h-0 md:overflow-hidden">
				<div class="flex items-center gap-3 mb-3 shrink-0">
					<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-muted-foreground">{m.today_schedule_label()}</span>
					<div class="flex-1 h-px bg-border"></div>
					<Button size="sm" variant="ghost" class="h-6 px-2 text-xs text-muted-foreground" onclick={() => createDialog?.open('event', new Date())}>
						<Plus class="w-3 h-3 mr-0.5" />{m.today_add_event_button()}
					</Button>
				</div>

				<!-- Mobile: flat list -->
				<div class="md:hidden flex flex-col gap-1">
					{#if sortedEvents.length > 0}
						{#each sortedEvents as event (event.id)}
							<button
								onclick={() => editDialog?.openEvent(event)}
								class="flex items-baseline gap-3 text-left py-1.5 px-2 -mx-2 rounded-md hover:bg-accent/50 transition-colors"
							>
								<span class="text-xs text-muted-foreground tabular-nums w-12 shrink-0 text-right">
									{event.all_day ? m.today_all_day() : fmtTime(event.start_at)}
								</span>
								<span class="text-sm font-medium">{event.title}</span>
								{#if event.location}
									<span class="text-xs text-muted-foreground truncate">{event.location}</span>
								{/if}
							</button>
						{/each}
					{:else}
						<p class="text-sm text-muted-foreground/50 italic">{m.today_nothing_scheduled()}</p>
					{/if}
				</div>

				<!-- Desktop: day calendar -->
				<div class="hidden md:block md:flex-1 md:min-h-0 md:overflow-hidden">
					<DayView
						{events}
						{tasks}
						{categories}
						onEventClick={(e, je) => showPreview('event', e, je)}
						onTaskClick={(t, je) => showPreview('task', t, je)}
						onDateClick={(date, allDay) => createDialog?.open('task', date, date, allDay)}
						onSelect={(start, end, allDay) => createDialog?.open('event', start, end, allDay)}
						onEventDrop={handleEventDrop}
						onTaskDrop={handleTaskDrop}
					/>
				</div>
			</div>

		</div>
	</div>
</div>

<CreateDialog bind:this={createDialog} {familyID} {members} {categories} onCreated={loadData} />
<EditDialog bind:this={editDialog} {familyID} {members} {categories} onSaved={loadData} onDeleted={loadData} />

<Popover.Root bind:open={previewOpen}>
	<Popover.Content customAnchor={previewAnchor} class="w-80 relative">
		{#if previewItem}
			{@const openEdit = () => {
				previewOpen = false;
				previewItem!.kind === 'event' ? editDialog?.openEvent(previewItem!.data as CalEvent) : editDialog?.openTask(previewItem!.data as Task);
			}}
			<div class="absolute top-2 right-2 flex items-center gap-0.5">
				<Button variant="ghost" size="icon" class="size-7 text-muted-foreground hover:text-foreground" onclick={openEdit} aria-label={m.action_edit()}>
					<Pencil class="size-3.5" />
				</Button>
				<Button
					variant="ghost"
					size="icon"
					class="size-7 text-muted-foreground hover:text-destructive"
					aria-label={m.edit_dialog_delete()}
					onclick={() => {
						previewOpen = false;
						previewItem!.kind === 'event' ? editDialog?.deleteEvent(previewItem!.data as CalEvent) : editDialog?.deleteTask(previewItem!.data as Task);
					}}
				>
					<Trash2 class="size-3.5" />
				</Button>
			</div>
			<div class="pr-16">
				{#if previewItem.kind === 'event'}
					<EventCard event={previewItem.data} {members} {categories} now={new Date()} interactive={false} />
				{:else}
					<TaskCard
						task={previewItem.data}
						{members}
						{categories}
						isDoneFilter={false}
						interactive={false}
						ontoggle={(e) => toggleTask(previewItem!.data as Task, e)}
					/>
				{/if}
			</div>
		{/if}
	</Popover.Content>
</Popover.Root>
