<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Checkbox } from '$lib/components/ui/checkbox';

	type Task = {
		id: string;
		title: string;
		status: string;
		start_date?: string;
		end_date?: string;
		created_by: string;
	};

	const familyID = $derived($page.params.id);

	let tasks = $state<Task[]>([]);
	let error = $state('');
	let newTaskTitle = $state('');
	let newTaskStart = $state('');
	let newTaskEnd = $state('');
	let addingTask = $state(false);

	let es: EventSource | null = null;

	onMount(async () => {
		try {
			tasks = (await api.get<Task[]>(`/api/v1/families/${familyID}/tasks`)) ?? [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load tasks';
		}

		es = new EventSource(sseUrl(`/api/v1/families/${familyID}/stream`));
		es.onmessage = async (e) => {
			if (e.data === 'refresh') {
				tasks = (await api.get<Task[]>(`/api/v1/families/${familyID}/tasks`)) ?? [];
			}
		};
		es.onerror = () => { es?.close(); es = null; };
	});

	onDestroy(() => es?.close());

	async function addTask() {
		if (!newTaskTitle.trim()) return;
		addingTask = true;
		try {
			await api.post<Task>(`/api/v1/families/${familyID}/tasks`, {
				title: newTaskTitle.trim(),
				start_date: newTaskStart ? new Date(newTaskStart).toISOString() : undefined,
				end_date: newTaskEnd ? new Date(newTaskEnd).toISOString() : undefined
			});
			newTaskTitle = '';
			newTaskStart = '';
			newTaskEnd = '';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create task';
		} finally {
			addingTask = false;
		}
	}

	async function toggleTask(task: Task) {
		const newStatus = task.status === 'done' ? 'todo' : 'done';
		try {
			await api.patch(`/api/v1/families/${familyID}/tasks/${task.id}`, {
				title: task.title,
				status: newStatus,
				start_date: task.start_date,
				end_date: task.end_date
			});
			tasks = tasks.map(t => t.id === task.id ? { ...t, status: newStatus } : t);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update task';
		}
	}

	async function deleteTask(taskID: string) {
		try {
			await api.delete(`/api/v1/families/${familyID}/tasks/${taskID}`);
			tasks = tasks.filter(t => t.id !== taskID);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete task';
		}
	}

	const todo = $derived(tasks.filter(t => t.status !== 'done'));
	const done = $derived(tasks.filter(t => t.status === 'done'));
</script>

{#if error}
	<p class="text-sm text-destructive">{error}</p>
{/if}

<form class="flex flex-col gap-2" onsubmit={(e) => { e.preventDefault(); addTask(); }}>
	<Input placeholder="New task…" bind:value={newTaskTitle} />
	<div class="flex gap-2">
		<Input type="date" bind:value={newTaskStart} class="flex-1" placeholder="Start date" />
		<Input type="date" bind:value={newTaskEnd} class="flex-1" placeholder="End date" />
		<Button type="submit" disabled={addingTask || !newTaskTitle.trim()}>Add</Button>
	</div>
</form>

{#if tasks.length === 0}
	<p class="text-sm text-muted-foreground">No tasks yet.</p>
{:else}
	<div class="flex flex-col gap-3">
		{#each todo as task (task.id)}
			<div class="flex items-center gap-3">
				<Checkbox checked={false} onCheckedChange={() => toggleTask(task)} id="task-{task.id}" />
				<div class="flex-1 min-w-0">
					<label for="task-{task.id}" class="text-sm cursor-pointer truncate block">{task.title}</label>
					{#if task.start_date || task.end_date}
						<p class="text-xs text-muted-foreground">
							{#if task.start_date && task.end_date}
								{new Date(task.start_date).toLocaleDateString()} → {new Date(task.end_date).toLocaleDateString()}
							{:else if task.end_date}
								Due {new Date(task.end_date).toLocaleDateString()}
							{:else if task.start_date}
								From {new Date(task.start_date).toLocaleDateString()}
							{/if}
						</p>
					{/if}
				</div>
				<Button variant="ghost" size="sm" class="text-muted-foreground hover:text-destructive h-7 px-2" onclick={() => deleteTask(task.id)}>✕</Button>
			</div>
		{/each}

		{#if done.length > 0}
			<p class="text-xs text-muted-foreground font-medium mt-2">Completed</p>
			{#each done as task (task.id)}
				<div class="flex items-center gap-3 opacity-50">
					<Checkbox checked={true} onCheckedChange={() => toggleTask(task)} id="task-{task.id}" />
					<div class="flex-1 min-w-0">
						<label for="task-{task.id}" class="text-sm cursor-pointer line-through truncate block">{task.title}</label>
					</div>
					<Button variant="ghost" size="sm" class="text-muted-foreground hover:text-destructive h-7 px-2" onclick={() => deleteTask(task.id)}>✕</Button>
				</div>
			{/each}
		{/if}
	</div>
{/if}
