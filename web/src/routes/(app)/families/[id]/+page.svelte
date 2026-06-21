<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { GripVertical, Plus } from 'lucide-svelte';
	import type { Task, CalEvent, Member, AppCategory } from '$lib/types';
	import { localDayMs, fmtTime } from '$lib/dates';
	import { sortable } from '$lib/sortable';
	import TaskCard from '$lib/components/TaskCard.svelte';
	import CreateDialog from '$lib/components/CreateDialog.svelte';
	import EditDialog from '$lib/components/EditDialog.svelte';

	const familyID = $derived($page.params.id ?? '');
	const now = new Date();
	const todayMs = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
	const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
	const todayEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999);

	let members = $state<Member[]>([]);
	let tasks = $state<Task[]>([]);
	let events = $state<CalEvent[]>([]);
	let categories = $state<AppCategory[]>([]);

	let createDialog: { open: (t?: 'task' | 'event') => void } | undefined = $state();
	let editDialog: { openTask: (t: Task) => void; openEvent: (e: CalEvent) => void } | undefined = $state();

	let es: EventSource | null = null;

	async function loadData() {
		const [membersRes, tasksRes, eventsRes, catsRes] = await Promise.allSettled([
			api.get<Member[]>(`/api/v1/families/${familyID}/members`),
			api.get<Task[]>(`/api/v1/families/${familyID}/tasks`),
			api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${todayStart.toISOString()}&to=${todayEnd.toISOString()}`),
			api.get<AppCategory[]>(`/api/v1/families/${familyID}/categories`),
		]);
		if (membersRes.status === 'fulfilled') members = membersRes.value ?? [];
		if (tasksRes.status === 'fulfilled') tasks = tasksRes.value ?? [];
		if (eventsRes.status === 'fulfilled') events = eventsRes.value ?? [];
		if (catsRes.status === 'fulfilled') categories = catsRes.value ?? [];
	}

	onMount(() => {
		loadData();
		es = new EventSource(sseUrl(`/api/v1/families/${familyID}/stream`));
		es.onmessage = (e) => { if (e.data === 'refresh') loadData(); };
		es.onerror = () => { es?.close(); es = null; };
	});
	onDestroy(() => es?.close());

	async function toggleTask(task: Task, e: MouseEvent) {
		e.stopPropagation();
		const newStatus = task.status === 'done' ? 'todo' : 'done';
		try {
			await api.patch(`/api/v1/families/${familyID}/tasks/${task.id}`, {
				title: task.title, description: task.description, important: task.important,
				status: newStatus, assigned_to: task.assigned_to, end_date: task.end_date, category_id: task.category_id,
			});
			tasks = tasks.map((t) => t.id === task.id ? { ...t, status: newStatus } : t);
		} catch { }
	}

	const sortedEvents = $derived(
		[...events].sort((a, b) => {
			if (a.all_day !== b.all_day) return a.all_day ? -1 : 1;
			return new Date(a.start_at).getTime() - new Date(b.start_at).getTime();
		}),
	);

	const overdueTasks = $derived(
		tasks
			.filter((t) => t.status !== 'done' && t.end_date && localDayMs(t.end_date) < todayMs)
			.sort((a, b) => new Date(a.end_date!).getTime() - new Date(b.end_date!).getTime()),
	);

	let dueTodayTasks = $state<Task[]>([]);
	$effect(() => {
		dueTodayTasks = tasks.filter((t) => t.status !== 'done' && t.end_date && localDayMs(t.end_date) === todayMs);
	});

	async function reorderTodayTasks(ids: string[]) {
		const prev = [...dueTodayTasks];
		dueTodayTasks = ids.map((id) => dueTodayTasks.find((t) => t.id === id)!).filter(Boolean);
		try {
			await api.put(`/api/v1/families/${familyID}/tasks/reorder`, { ids });
		} catch { dueTodayTasks = prev; }
	}
</script>

<!-- Header -->
<div class="sticky top-0 z-10 bg-background px-4 md:px-6 pt-4 md:pt-6 pb-3">
	<h1 class="text-xl font-semibold">Today</h1>
	<p class="text-xs text-muted-foreground">
		{now.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })}
	</p>
</div>

<div class="px-4 md:px-6 flex flex-col gap-6 pb-8">
	<!-- Schedule -->
	<div>
		<div class="flex items-center gap-3 mb-3">
			<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-muted-foreground">Schedule</span>
			<div class="flex-1 h-px bg-border"></div>
			<Button size="sm" variant="ghost" class="h-6 px-2 text-xs text-muted-foreground" onclick={() => createDialog?.open('event')}>
				<Plus class="w-3 h-3 mr-0.5" />Event
			</Button>
		</div>
		{#if sortedEvents.length > 0}
			<div class="flex flex-col gap-1">
				{#each sortedEvents as event (event.id)}
					<button
						onclick={() => editDialog?.openEvent(event)}
						class="flex items-baseline gap-3 text-left py-1.5 px-2 -mx-2 rounded-md hover:bg-accent/50 transition-colors cursor-pointer"
					>
						<span class="text-xs text-muted-foreground tabular-nums w-12 shrink-0 text-right">
							{event.all_day ? 'All day' : fmtTime(event.start_at)}
						</span>
						<span class="text-sm font-medium">{#if event.icon}<span class="mr-1">{event.icon}</span>{/if}{event.title}</span>
						{#if event.location}
							<span class="text-xs text-muted-foreground truncate hidden sm:block">{event.location}</span>
						{/if}
					</button>
				{/each}
			</div>
		{:else}
			<p class="text-sm text-muted-foreground/50 italic">Nothing scheduled</p>
		{/if}
	</div>

	<!-- Overdue -->
	{#if overdueTasks.length > 0}
		<div>
			<div class="flex items-center gap-3 mb-2">
				<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-destructive">Overdue</span>
				<div class="flex-1 h-px bg-destructive/20"></div>
			</div>
			<div class="flex flex-col gap-2">
				{#each overdueTasks as task (task.id)}
					<TaskCard {task} {members} {categories} isDoneFilter={false}
						onclick={() => editDialog?.openTask(task)}
						ontoggle={(e) => toggleTask(task, e)} />
				{/each}
			</div>
		</div>
	{/if}

	<!-- Due today -->
	<div>
		<div class="flex items-center gap-3 mb-2">
			<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-foreground">Due today</span>
			<div class="flex-1 h-px bg-border"></div>
			<Button size="sm" variant="ghost" class="h-6 px-2 text-xs text-muted-foreground" onclick={() => createDialog?.open('task')}>
				<Plus class="w-3 h-3 mr-0.5" />Task
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
			<p class="text-sm text-muted-foreground/50 italic">No tasks due today</p>
		{/if}
	</div>
</div>

<CreateDialog bind:this={createDialog} {familyID} {members} {categories} onCreated={loadData} />
<EditDialog bind:this={editDialog} {familyID} {members} {categories} onSaved={loadData} onDeleted={loadData} />
