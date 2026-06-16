<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';

	type CalEvent = {
		id: string;
		title: string;
		start_at: string;
		end_at: string;
		all_day: boolean;
		description?: string;
	};

	type Task = { id: string; title: string; start_date?: string; end_date?: string; status: string };

	const familyID = $derived($page.params.id);

	let today = new Date();
	let viewYear = $state(today.getFullYear());
	let viewMonth = $state(today.getMonth()); // 0-indexed

	let events = $state<CalEvent[]>([]);
	let tasks = $state<Task[]>([]);
	let error = $state('');

	// Create/edit dialog state
	let showForm = $state(false);
	let editing = $state<CalEvent | null>(null);
	let formTitle = $state('');
	let formStart = $state('');
	let formEnd = $state('');
	let formAllDay = $state(false);
	let formDesc = $state('');
	let saving = $state(false);

	async function loadMonth() {
		const from = new Date(viewYear, viewMonth, 1);
		const to = new Date(viewYear, viewMonth + 1, 1);
		try {
			[events, tasks] = await Promise.all([
				api.get<CalEvent[]>(
					`/api/v1/families/${familyID}/events?from=${from.toISOString()}&to=${to.toISOString()}`
				).then(r => r ?? []),
				api.get<Task[]>(`/api/v1/families/${familyID}/tasks`).then(r => r ?? [])
			]);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load calendar';
		}
	}

	let es: EventSource | null = null;

	onMount(() => {
		loadMonth();
		es = new EventSource(sseUrl(`/api/v1/families/${familyID}/stream`));
		es.onmessage = (e) => { if (e.data === 'refresh') loadMonth(); };
		es.onerror = () => { es?.close(); es = null; };
	});

	onDestroy(() => es?.close());

	function prevMonth() {
		if (viewMonth === 0) { viewYear--; viewMonth = 11; } else { viewMonth--; }
		loadMonth();
	}

	function nextMonth() {
		if (viewMonth === 11) { viewYear++; viewMonth = 0; } else { viewMonth++; }
		loadMonth();
	}

	// Build the 6-row grid of days
	const calDays = $derived((() => {
		const first = new Date(viewYear, viewMonth, 1);
		const last = new Date(viewYear, viewMonth + 1, 0);
		const startPad = first.getDay(); // 0=Sun
		const days: (Date | null)[] = [];
		for (let i = 0; i < startPad; i++) days.push(null);
		for (let d = 1; d <= last.getDate(); d++) days.push(new Date(viewYear, viewMonth, d));
		while (days.length % 7 !== 0) days.push(null);
		return days;
	})());

	function eventsForDay(date: Date): CalEvent[] {
		return events.filter(e => {
			const s = new Date(e.start_at);
			return s.getFullYear() === date.getFullYear() &&
				s.getMonth() === date.getMonth() &&
				s.getDate() === date.getDate();
		});
	}

	type TaskSpan = { task: Task; isStart: boolean; isEnd: boolean };

	function taskSpansForDay(date: Date): TaskSpan[] {
		const dayStart = new Date(date); dayStart.setHours(0, 0, 0, 0);
		const dayEnd = new Date(date); dayEnd.setHours(23, 59, 59, 999);
		return tasks
			.filter(t => {
				if (t.status === 'done' || (!t.start_date && !t.end_date)) return false;
				const s = t.start_date ? new Date(t.start_date) : new Date(t.end_date!);
				const e = t.end_date ? new Date(t.end_date) : new Date(t.start_date!);
				return s <= dayEnd && e >= dayStart;
			})
			.map(t => {
				const s = t.start_date ? new Date(t.start_date) : new Date(t.end_date!);
				const e = t.end_date ? new Date(t.end_date) : new Date(t.start_date!);
				// treat week boundaries as visual start/end so bars wrap cleanly
				const isStart = s >= dayStart || date.getDay() === 0;
				const isEnd   = e <= dayEnd   || date.getDay() === 6;
				return { task: t, isStart, isEnd };
			});
	}

	function isToday(date: Date) {
		return date.toDateString() === today.toDateString();
	}

	function openNew(date?: Date) {
		editing = null;
		formTitle = '';
		formDesc = '';
		formAllDay = false;
		if (date) {
			const iso = date.toISOString().slice(0, 10);
			formStart = `${iso}T09:00`;
			formEnd = `${iso}T10:00`;
		} else {
			formStart = '';
			formEnd = '';
		}
		showForm = true;
	}

	function openEdit(e: CalEvent) {
		editing = e;
		formTitle = e.title;
		formDesc = e.description ?? '';
		formAllDay = e.all_day;
		formStart = e.start_at.slice(0, 16);
		formEnd = e.end_at.slice(0, 16);
		showForm = true;
	}

	async function saveEvent() {
		if (!formTitle.trim() || !formStart || !formEnd) return;
		saving = true;
		try {
			const body = {
				title: formTitle.trim(),
				description: formDesc,
				start_at: new Date(formStart).toISOString(),
				end_at: new Date(formEnd).toISOString(),
				all_day: formAllDay
			};
			if (editing) {
				await api.patch(`/api/v1/families/${familyID}/events/${editing.id}`, body);
			} else {
				await api.post(`/api/v1/families/${familyID}/events`, body);
			}
			showForm = false;
			await loadMonth();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save event';
		} finally {
			saving = false;
		}
	}

	async function deleteEvent() {
		if (!editing) return;
		try {
			await api.delete(`/api/v1/families/${familyID}/events/${editing.id}`);
			showForm = false;
			await loadMonth();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete event';
		}
	}

	const monthLabel = $derived(
		new Date(viewYear, viewMonth, 1).toLocaleString('default', { month: 'long', year: 'numeric' })
	);

	const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
</script>

{#if error}
	<p class="text-sm text-destructive mb-2">{error}</p>
{/if}

<!-- Header -->
<div class="flex items-center justify-between mb-3">
	<div class="flex items-center gap-2">
		<Button variant="outline" size="sm" onclick={prevMonth}>‹</Button>
		<span class="text-sm font-medium w-36 text-center">{monthLabel}</span>
		<Button variant="outline" size="sm" onclick={nextMonth}>›</Button>
	</div>
	<Button size="sm" onclick={() => openNew()}>+ Event</Button>
</div>

<!-- Grid -->
<div class="grid grid-cols-7 border-l border-t rounded-lg overflow-hidden text-xs">
	{#each DAYS as d}
		<div class="border-r border-b px-1 py-1.5 text-center text-muted-foreground font-semibold bg-muted/40 text-[11px] uppercase tracking-wide">{d}</div>
	{/each}
	{#each calDays as day}
		<div
			class="border-r border-b min-h-24 flex flex-col cursor-pointer hover:bg-muted/20 transition-colors group"
			onclick={() => day && openNew(day)}
			role="button"
			tabindex="0"
			onkeydown={(e) => e.key === 'Enter' && day && openNew(day)}
		>
			{#if day}
				<!-- date number -->
				<div class="px-1.5 pt-1.5 pb-1">
					<span class="text-xs font-medium inline-flex items-center justify-center w-6 h-6 rounded-full
						{isToday(day)
							? 'bg-primary text-primary-foreground font-bold'
							: 'text-foreground group-hover:bg-muted/60'}">
						{day.getDate()}
					</span>
				</div>

				<!-- calendar events (single-day chips) -->
				<div class="flex flex-col gap-0.5 px-1 pb-0.5">
					{#each eventsForDay(day) as ev}
						<button
							class="w-full text-left truncate rounded-md px-1.5 py-0.5 bg-blue-500/15 text-blue-700 dark:text-blue-300 text-[11px] font-medium leading-5 hover:bg-blue-500/25 transition-colors"
							onclick={(e) => { e.stopPropagation(); openEdit(ev); }}
						>
							{ev.title}
						</button>
					{/each}
				</div>

				<!-- task spanning bars -->
				<div class="flex flex-col gap-0.5 pb-1.5">
					{#each taskSpansForDay(day) as { task, isStart, isEnd }}
						<div class="
							text-[11px] leading-5 h-5 bg-primary/20 text-primary overflow-hidden font-medium
							{isStart ? 'rounded-l-md ml-1 pl-1.5' : '-ml-px pl-0'}
							{isEnd   ? 'rounded-r-md mr-1 pr-1'   : '-mr-px'}
						">
							{#if isStart}
								<span class="truncate block">{task.title}</span>
							{:else}
								&nbsp;
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/each}
</div>

<!-- Event form dialog -->
{#if showForm}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={() => showForm = false}>
		<div class="bg-background border rounded-lg p-6 w-full max-w-sm flex flex-col gap-4" onclick={(e) => e.stopPropagation()}>
			<h3 class="font-semibold">{editing ? 'Edit event' : 'New event'}</h3>

			<div class="flex flex-col gap-1">
				<Label for="ev-title">Title</Label>
				<Input id="ev-title" bind:value={formTitle} placeholder="Event title" />
			</div>

			<div class="flex flex-col gap-1">
				<Label for="ev-start">Start</Label>
				<Input id="ev-start" type="datetime-local" bind:value={formStart} />
			</div>

			<div class="flex flex-col gap-1">
				<Label for="ev-end">End</Label>
				<Input id="ev-end" type="datetime-local" bind:value={formEnd} />
			</div>

			<div class="flex flex-col gap-1">
				<Label for="ev-desc">Description (optional)</Label>
				<Textarea id="ev-desc" bind:value={formDesc} placeholder="Add a description…" rows={3} />
			</div>

			<div class="flex gap-2 justify-end">
				{#if editing}
					<Button variant="destructive" size="sm" onclick={deleteEvent}>Delete</Button>
				{/if}
				<Button variant="outline" size="sm" onclick={() => showForm = false}>Cancel</Button>
				<Button size="sm" onclick={saveEvent} disabled={saving || !formTitle.trim()}>
					{saving ? 'Saving…' : 'Save'}
				</Button>
			</div>
		</div>
	</div>
{/if}
