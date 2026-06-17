<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { CheckSquare } from 'lucide-svelte';
	import type { Task, CalEvent, Member, AppLabel, Filter } from '$lib/types';
	import BoardToolbar from '$lib/components/BoardToolbar.svelte';
	import TaskCard from '$lib/components/TaskCard.svelte';
	import EventCard from '$lib/components/EventCard.svelte';
	import CreateDialog from '$lib/components/CreateDialog.svelte';
	import EditDialog from '$lib/components/EditDialog.svelte';

	const familyID = $derived($page.params.id ?? '');
	const now = new Date();

	let members = $state<Member[]>([]);
	let tasks = $state<Task[]>([]);
	let events = $state<CalEvent[]>([]);
	let labels = $state<AppLabel[]>([]);
	let error = $state('');
	let filter = $state<Filter>('all');
	let filterMembers = $state(new Set<string>());
	let filterLabels = $state(new Set<string>());
	let sortBy = $state<'date' | 'priority' | 'title'>('date');
	let sortAsc = $state(true);

	let createDialog: { open: (t?: 'task' | 'event') => void } | undefined = $state();
	let editDialog: { openTask: (t: Task) => void; openEvent: (e: CalEvent) => void } | undefined = $state();

	let es: EventSource | null = null;
	let errorTimer: ReturnType<typeof setTimeout> | null = null;

	async function loadData() {
		const from = new Date();
		const to = new Date();
		to.setDate(to.getDate() + 90);
		const [membersRes, tasksRes, eventsRes, labelsRes] = await Promise.allSettled([
			api.get<Member[]>(`/api/v1/families/${familyID}/members`),
			api.get<Task[]>(`/api/v1/families/${familyID}/tasks`),
			api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${from.toISOString()}&to=${to.toISOString()}`),
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

	// ── Filtered + sorted list ────────────────────────────────────────────────
	const PRIORITY_RANK: Record<string, number> = { high: 0, medium: 1 };

	type ListItem =
		| { kind: 'task'; data: Task; sortKey: number }
		| { kind: 'event'; data: CalEvent; sortKey: number };

	const visibleItems = $derived(
		(() => {
			const items: ListItem[] = [];
			const FAR_FUTURE = 9999999999999;
			const matchesMember = (t: Task) =>
				filterMembers.size === 0 || (!!t.assigned_to && filterMembers.has(t.assigned_to));
			const matchesMemberEv = (ev: CalEvent) =>
				filterMembers.size === 0 || (ev.attendee_ids ?? []).some((id) => filterMembers.has(id));
			const matchesLabel = (ids: string[] | undefined) =>
				filterLabels.size === 0 || (ids ?? []).some((id) => filterLabels.has(id));

			if (filter === 'all' || filter === 'tasks') {
				for (const t of tasks.filter((t) => t.status !== 'done' && matchesMember(t) && matchesLabel(t.label_ids))) {
					const sortKey = t.end_date
						? new Date(t.end_date).getTime()
						: FAR_FUTURE + (PRIORITY_RANK[t.priority] ?? 1);
					items.push({ kind: 'task', data: t, sortKey });
				}
			}
			if (filter === 'all' || filter === 'events') {
				for (const ev of events.filter((ev) => new Date(ev.end_at) >= now && matchesMemberEv(ev) && matchesLabel(ev.label_ids))) {
					items.push({ kind: 'event', data: ev, sortKey: new Date(ev.start_at).getTime() });
				}
			}
			if (filter === 'done') {
				for (const t of tasks.filter((t) => t.status === 'done' && matchesMember(t))) {
					items.push({ kind: 'task', data: t, sortKey: 0 });
				}
			}

			const dir = sortAsc ? 1 : -1;
			if (sortBy === 'date') return items.sort((a, b) => (a.sortKey - b.sortKey) * dir);
			if (sortBy === 'priority')
				return items.sort((a, b) => {
					const pa = a.kind === 'task' ? (PRIORITY_RANK[a.data.priority] ?? 1) : 3;
					const pb = b.kind === 'task' ? (PRIORITY_RANK[b.data.priority] ?? 1) : 3;
					return (pa !== pb ? pa - pb : a.sortKey - b.sortKey) * dir;
				});
			return items.sort((a, b) => a.data.title.localeCompare(b.data.title) * dir);
		})(),
	);

	const doneCnt = $derived(tasks.filter((t) => t.status === 'done').length);
</script>

{#if error}
	<div class="flex items-center justify-between gap-2 px-3 py-2 mb-3 rounded-md bg-destructive/10 text-destructive text-sm">
		<span>{error}</span>
		<button onclick={() => (error = '')} class="shrink-0 opacity-70 hover:opacity-100">✕</button>
	</div>
{/if}

<div class="flex items-center justify-between mb-4">
	<h1 class="text-xl font-semibold">Board</h1>
	<Button onclick={() => createDialog?.open(filter === 'events' ? 'event' : 'task')}>+ New</Button>
</div>

<BoardToolbar
	bind:filter
	bind:filterMembers
	bind:filterLabels
	bind:sortBy
	bind:sortAsc
	{members}
	{labels}
	{doneCnt}
	{familyID}
/>

{#if visibleItems.length === 0}
	<div class="flex flex-col items-center gap-2 py-16 text-muted-foreground">
		<CheckSquare class="w-10 h-10 opacity-30" />
		<p class="text-sm">
			{filter === 'done' ? 'Nothing completed yet.' : 'Nothing here yet. Add a task or event above.'}
		</p>
	</div>
{:else}
	<div class="flex flex-col gap-2">
		{#each visibleItems as item (item.kind + item.data.id)}
			{#if item.kind === 'task'}
				<TaskCard
					task={item.data}
					{members}
					{labels}
					isDoneFilter={filter === 'done'}
					onclick={() => editDialog?.openTask(item.data)}
					ontoggle={(e) => toggleTask(item.data, e)}
				/>
			{:else}
				<EventCard
					event={item.data}
					{members}
					{labels}
					{now}
					onclick={() => editDialog?.openEvent(item.data)}
				/>
			{/if}
		{/each}
	</div>
{/if}

<CreateDialog
	bind:this={createDialog}
	{familyID}
	{members}
	{labels}
	onCreated={loadData}
	onError={setError}
/>

<EditDialog
	bind:this={editDialog}
	{familyID}
	{members}
	{labels}
	onSaved={loadData}
	onDeleted={loadData}
	onError={setError}
/>
