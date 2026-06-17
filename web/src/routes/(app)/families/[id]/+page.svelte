<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { CalendarDays } from 'lucide-svelte';
	import type { Task, CalEvent, Member, AppLabel } from '$lib/types';
	import { localDayMs, fmtDateTime } from '$lib/dates';
	import TaskCard from '$lib/components/TaskCard.svelte';
	import EventCard from '$lib/components/EventCard.svelte';
	import CreateDialog from '$lib/components/CreateDialog.svelte';
	import EditDialog from '$lib/components/EditDialog.svelte';

	const familyID = $derived($page.params.id ?? '');
	const now = new Date();
	const todayMs = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();

	let members = $state<Member[]>([]);
	let tasks = $state<Task[]>([]);
	let events = $state<CalEvent[]>([]);
	let labels = $state<AppLabel[]>([]);
	let error = $state('');
	let quickTitle = $state('');

	let createDialog: { open: (t?: 'task' | 'event') => void } | undefined = $state();
	let editDialog: { openTask: (t: Task) => void; openEvent: (e: CalEvent) => void } | undefined = $state();

	let es: EventSource | null = null;
	let errorTimer: ReturnType<typeof setTimeout> | null = null;

	async function loadData() {
		const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
		const todayEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999);
		const [membersRes, tasksRes, eventsRes, labelsRes] = await Promise.allSettled([
			api.get<Member[]>(`/api/v1/families/${familyID}/members`),
			api.get<Task[]>(`/api/v1/families/${familyID}/tasks`),
			api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${todayStart.toISOString()}&to=${todayEnd.toISOString()}`),
			api.get<AppLabel[]>(`/api/v1/families/${familyID}/labels`),
		]);
		if (membersRes.status === 'fulfilled') members = membersRes.value ?? [];
		else setError(membersRes.reason);
		if (tasksRes.status === 'fulfilled') tasks = tasksRes.value ?? [];
		else setError(tasksRes.reason);
		if (eventsRes.status === 'fulfilled') events = eventsRes.value ?? [];
		else setError(eventsRes.reason);
		if (labelsRes.status === 'fulfilled') labels = labelsRes.value ?? [];
		else setError(labelsRes.reason);
	}

	onMount(() => {
		loadData();
		es = new EventSource(sseUrl(`/api/v1/families/${familyID}/stream`));
		es.onmessage = (e) => { if (e.data === 'refresh') loadData(); };
		es.onerror = () => { es?.close(); es = null; };
	});

	onDestroy(() => {
		es?.close();
		if (errorTimer) clearTimeout(errorTimer);
	});

	function setError(err: unknown) {
		error = err instanceof Error ? err.message : 'Something went wrong';
		if (errorTimer) clearTimeout(errorTimer);
		errorTimer = setTimeout(() => (error = ''), 4000);
	}

	async function toggleTask(task: Task, e: MouseEvent) {
		e.stopPropagation();
		const newStatus = task.status === 'done' ? 'todo' : 'done';
		try {
			await api.patch(`/api/v1/families/${familyID}/tasks/${task.id}`, {
				title: task.title, description: task.description,
				priority: task.priority, status: newStatus,
				assigned_to: task.assigned_to, end_date: task.end_date,
				label_ids: task.label_ids ?? [],
			});
			tasks = tasks.map((t) => (t.id === task.id ? { ...t, status: newStatus } : t));
		} catch (err) {
			setError(err);
		}
	}

	async function quickAdd(e: SubmitEvent) {
		e.preventDefault();
		if (!quickTitle.trim()) return;
		try {
			await api.post(`/api/v1/families/${familyID}/tasks`, {
				title: quickTitle.trim(), priority: 'medium', label_ids: [],
			});
			quickTitle = '';
			loadData();
		} catch (err) {
			setError(err);
		}
	}

	const PRIORITY_RANK: Record<string, number> = { high: 0, medium: 1 };

	const overdueTasks = $derived(
		tasks
			.filter((t) => t.status !== 'done' && t.end_date && localDayMs(t.end_date) < todayMs)
			.sort((a, b) => new Date(a.end_date!).getTime() - new Date(b.end_date!).getTime()),
	);

	const dueTodayTasks = $derived(
		tasks
			.filter((t) => t.status !== 'done' && t.end_date && localDayMs(t.end_date) === todayMs)
			.sort((a, b) => (PRIORITY_RANK[a.priority] ?? 2) - (PRIORITY_RANK[b.priority] ?? 2)),
	);

	const todayEvents = $derived(
		[...events].sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime()),
	);

	const isEmpty = $derived(overdueTasks.length === 0 && dueTodayTasks.length === 0 && todayEvents.length === 0);
</script>

{#if error}
	<div class="flex items-center justify-between gap-2 px-3 py-2 mb-3 rounded-md bg-destructive/10 text-destructive text-sm">
		<span>{error}</span>
		<button onclick={() => (error = '')} class="shrink-0 opacity-70 hover:opacity-100">✕</button>
	</div>
{/if}

<div class="flex items-center justify-between mb-4">
	<div>
		<h1 class="text-xl font-semibold">Today</h1>
		<p class="text-xs text-muted-foreground">{now.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })}</p>
	</div>
	<Button size="sm" onclick={() => createDialog?.open('event')}>
		<CalendarDays class="w-3.5 h-3.5 mr-1" />Event
	</Button>
</div>

<form onsubmit={quickAdd} class="mb-5">
	<Input bind:value={quickTitle} placeholder="Add a task for today…" class="bg-muted/20 border-dashed focus-visible:border-solid" />
</form>

{#if isEmpty}
	<div class="flex flex-col items-center gap-2 py-16 text-muted-foreground">
		<p class="text-2xl">✓</p>
		<p class="text-sm font-medium">All clear for today</p>
		<p class="text-xs">No events, no tasks due.</p>
	</div>
{:else}
	{#if overdueTasks.length > 0}
		<div class="flex items-center gap-3 mb-2">
			<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-destructive">Overdue</span>
			<div class="flex-1 h-px bg-destructive/20"></div>
		</div>
		<div class="flex flex-col gap-2 mb-5">
			{#each overdueTasks as task (task.id)}
				<TaskCard {task} {members} {labels} isDoneFilter={false}
					onclick={() => editDialog?.openTask(task)}
					ontoggle={(e) => toggleTask(task, e)} />
			{/each}
		</div>
	{/if}

	{#if todayEvents.length > 0}
		<div class="flex items-center gap-3 mb-2">
			<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-foreground">Events</span>
			<div class="flex-1 h-px bg-border"></div>
		</div>
		<div class="flex flex-col gap-2 mb-5">
			{#each todayEvents as event (event.id)}
				<EventCard {event} {members} {labels} {now} onclick={() => editDialog?.openEvent(event)} />
			{/each}
		</div>
	{/if}

	{#if dueTodayTasks.length > 0}
		<div class="flex items-center gap-3 mb-2">
			<span class="text-xs font-semibold uppercase tracking-wide shrink-0 text-foreground">Due today</span>
			<div class="flex-1 h-px bg-border"></div>
		</div>
		<div class="flex flex-col gap-2">
			{#each dueTodayTasks as task (task.id)}
				<TaskCard {task} {members} {labels} isDoneFilter={false}
					onclick={() => editDialog?.openTask(task)}
					ontoggle={(e) => toggleTask(task, e)} />
			{/each}
		</div>
	{/if}
{/if}

<CreateDialog bind:this={createDialog} {familyID} {members} {labels} onCreated={loadData} onError={setError} />
<EditDialog bind:this={editDialog} {familyID} {members} {labels} onSaved={loadData} onDeleted={loadData} onError={setError} />
