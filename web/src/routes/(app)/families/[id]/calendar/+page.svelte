<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Popover from '$lib/components/ui/popover';
	import { RangeCalendar } from '$lib/components/ui/range-calendar';
	import { CalendarDate, type DateValue } from '@internationalized/date';
	import type { DateRange } from 'bits-ui';
	import { CalendarDays } from 'lucide-svelte';

	type CalEvent = {
		id: string;
		title: string;
		start_at: string;
		end_at: string;
		all_day: boolean;
		description?: string;
		location?: string;
		label_ids?: string[];
	};

	type Task = { id: string; title: string; start_date?: string; end_date?: string; status: string };

	type AppLabel = { id: string; family_id: string; name: string; color: string };

	type LabelColor = 'red' | 'orange' | 'yellow' | 'green' | 'teal' | 'blue' | 'purple' | 'pink' | 'gray';

	const LABEL_CHIP: Record<LabelColor, string> = {
		red: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
		orange: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
		yellow: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
		green: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
		teal: 'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400',
		blue: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
		purple: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
		pink: 'bg-pink-100 text-pink-700 dark:bg-pink-900/30 dark:text-pink-400',
		gray: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
	};

	const LABEL_DOT: Record<LabelColor, string> = {
		red: 'bg-red-500', orange: 'bg-orange-500', yellow: 'bg-yellow-500',
		green: 'bg-green-500', teal: 'bg-teal-500', blue: 'bg-blue-500',
		purple: 'bg-purple-500', pink: 'bg-pink-500', gray: 'bg-gray-400',
	};

	function chipClass(color: string) { return LABEL_CHIP[color as LabelColor] ?? LABEL_CHIP.gray; }
	function dotClass(color: string) { return LABEL_DOT[color as LabelColor] ?? LABEL_DOT.gray; }

	const familyID = $derived($page.params.id);

	let today = new Date();
	let viewYear = $state(today.getFullYear());
	let viewMonth = $state(today.getMonth());

	let events = $state<CalEvent[]>([]);
	let tasks = $state<Task[]>([]);
	let labels = $state<AppLabel[]>([]);
	let error = $state('');

	function labelByID(id: string) { return labels.find((l) => l.id === id); }

	let showForm = $state(false);
	let editing = $state<CalEvent | null>(null);
	let formTitle = $state('');
	let formAllDay = $state(false);
	let formLabelIDs = $state<string[]>([]);
	let formDesc = $state('');
	let formLocation = $state('');
	let saving = $state(false);
	let pickerOpen = $state(false);

	let dateRange = $state<DateRange>({ start: undefined, end: undefined });
	let startTime = $state('09:00');
	let endTime = $state('10:00');

	function toCalDate(isoStr: string): CalendarDate {
		const d = new Date(isoStr);
		return new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate());
	}

	function toISO(date: DateValue, time: string, allDay: boolean): string {
		const pad = (n: number) => String(n).padStart(2, '0');
		const dateStr = `${date.year}-${pad(date.month)}-${pad(date.day)}`;
		return new Date(`${dateStr}T${allDay ? '00:00' : time}`).toISOString();
	}

	const rangeLabel = $derived((() => {
		const { start, end } = dateRange;
		if (!start) return 'Select dates';
		const fmt = (d: DateValue) => `${d.month}/${d.day}/${d.year}`;
		if (!end || (end.day === start.day && end.month === start.month && end.year === start.year)) return fmt(start);
		return `${fmt(start)} – ${fmt(end)}`;
	})());

	async function loadMonth() {
		const from = new Date(viewYear, viewMonth, 1);
		const to = new Date(viewYear, viewMonth + 1, 1);
		try {
			[events, tasks, labels] = await Promise.all([
				api.get<CalEvent[]>(
					`/api/v1/families/${familyID}/events?from=${from.toISOString()}&to=${to.toISOString()}`
				).then(r => r ?? []),
				api.get<Task[]>(`/api/v1/families/${familyID}/tasks`).then(r => r ?? []),
				api.get<AppLabel[]>(`/api/v1/families/${familyID}/labels`).then(r => r ?? []),
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

	const calDays = $derived((() => {
		const first = new Date(viewYear, viewMonth, 1);
		const last = new Date(viewYear, viewMonth + 1, 0);
		const startPad = first.getDay();
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
		formLocation = '';
		formAllDay = false;
		formLabelIDs = [];
		startTime = '09:00';
		endTime = '10:00';
		if (date) {
			const cd = new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate());
			dateRange = { start: cd, end: cd };
		} else {
			dateRange = { start: undefined, end: undefined };
		}
		showForm = true;
	}

	function openEdit(e: CalEvent) {
		editing = e;
		formTitle = e.title;
		formDesc = e.description ?? '';
		formLocation = e.location ?? '';
		formAllDay = e.all_day;
		formLabelIDs = [...(e.label_ids ?? [])];
		const s = new Date(e.start_at);
		const en = new Date(e.end_at);
		dateRange = { start: toCalDate(e.start_at), end: toCalDate(e.end_at) };
		startTime = `${String(s.getHours()).padStart(2, '0')}:${String(s.getMinutes()).padStart(2, '0')}`;
		endTime = `${String(en.getHours()).padStart(2, '0')}:${String(en.getMinutes()).padStart(2, '0')}`;
		showForm = true;
	}

	async function saveEvent() {
		if (!formTitle.trim() || !dateRange.start) return;
		saving = true;
		const endDate = dateRange.end ?? dateRange.start;
		try {
			const body = {
				title: formTitle.trim(),
				description: formDesc,
				location: formLocation,
				start_at: toISO(dateRange.start, startTime, formAllDay),
				end_at: toISO(endDate, endTime, formAllDay),
				all_day: formAllDay,
				label_ids: formLabelIDs,
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
				<div class="px-1.5 pt-1.5 pb-1">
					<span class="text-xs font-medium inline-flex items-center justify-center w-6 h-6 rounded-full
						{isToday(day)
							? 'bg-primary text-primary-foreground font-bold'
							: 'text-foreground group-hover:bg-muted/60'}">
						{day.getDate()}
					</span>
				</div>

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
<Dialog.Root bind:open={showForm}>
	<Dialog.Portal>
		<Dialog.Overlay />
		<Dialog.Content class="sm:max-w-md">
			<Dialog.Header>
				<Dialog.Title>{editing ? 'Edit event' : 'New event'}</Dialog.Title>
			</Dialog.Header>

			<div class="flex flex-col gap-4 py-2">
				<div class="flex flex-col gap-1.5">
					<Label for="ev-title">Title</Label>
					<Input id="ev-title" bind:value={formTitle} placeholder="Event title" />
				</div>

				<!-- Date range picker -->
				<div class="flex flex-col gap-1.5">
					<Label>Dates</Label>
					<Popover.Root bind:open={pickerOpen}>
						<Popover.Trigger>
							<Button variant="outline" class="w-full justify-start gap-2 font-normal">
								<CalendarDays class="w-4 h-4 text-muted-foreground" />
								{rangeLabel}
							</Button>
						</Popover.Trigger>
						<Popover.Content class="w-auto p-0" align="start">
							<RangeCalendar
								bind:value={dateRange}
								onValueChange={() => { if (dateRange.start && dateRange.end) pickerOpen = false; }}
							/>
						</Popover.Content>
					</Popover.Root>
				</div>

				<!-- Time inputs (hidden when all-day) -->
				{#if !formAllDay}
					<div class="flex gap-3">
						<div class="flex flex-col gap-1.5 flex-1">
							<Label for="ev-start-time">Start time</Label>
							<Input id="ev-start-time" type="time" bind:value={startTime} />
						</div>
						<div class="flex flex-col gap-1.5 flex-1">
							<Label for="ev-end-time">End time</Label>
							<Input id="ev-end-time" type="time" bind:value={endTime} />
						</div>
					</div>
				{/if}

				<label class="flex items-center gap-2 text-sm cursor-pointer">
					<Checkbox bind:checked={formAllDay} />
					All day
				</label>

				<div class="flex flex-col gap-1.5">
					<Label for="ev-location">Location</Label>
					<Input id="ev-location" bind:value={formLocation} placeholder="Optional location…" />
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="ev-desc">Description</Label>
					<Textarea id="ev-desc" bind:value={formDesc} placeholder="Optional details…" rows={2} />
				</div>

				{#if labels.length > 0}
					<div class="flex flex-col gap-1.5">
						<Label>Labels</Label>
						<div class="flex flex-wrap gap-1.5">
							{#each labels as lbl}
								<button
									type="button"
									onclick={() => {
										formLabelIDs = formLabelIDs.includes(lbl.id)
											? formLabelIDs.filter((id) => id !== lbl.id)
											: [...formLabelIDs, lbl.id];
									}}
									class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-all border
										{formLabelIDs.includes(lbl.id)
										? 'border-foreground ring-1 ring-foreground ' + chipClass(lbl.color)
										: 'border-transparent ' + chipClass(lbl.color)}"
								>
									<span class="w-1.5 h-1.5 rounded-full {dotClass(lbl.color)} shrink-0"></span>
									{lbl.name}
								</button>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<Dialog.Footer class="flex-col-reverse sm:flex-row gap-2">
				{#if editing}
					<Button variant="destructive" onclick={deleteEvent}>Delete</Button>
				{/if}
				<div class="flex gap-2 sm:ml-auto">
					<Button variant="outline" onclick={() => (showForm = false)}>Cancel</Button>
					<Button onclick={saveEvent} disabled={saving || !formTitle.trim() || !dateRange.start}>
						{saving ? 'Saving…' : 'Save'}
					</Button>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
