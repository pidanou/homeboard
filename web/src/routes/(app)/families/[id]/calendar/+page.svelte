<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy, tick } from 'svelte';
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
	import { CalendarDays, Filter, SquareCheckBig } from 'lucide-svelte';
	import type { CalEvent, Task, Member, AppLabel } from '$lib/types';

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
	const today = new Date();
	const todayMs = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime();

	// View
	let view = $state<'month' | 'agenda'>('month');

	// Month nav
	let viewYear = $state(today.getFullYear());
	let viewMonth = $state(today.getMonth());

	// Data
	let events = $state<CalEvent[]>([]);
	let tasks = $state<Task[]>([]);
	let members = $state<Member[]>([]);
	let labels = $state<AppLabel[]>([]);
	let error = $state('');

	// Filters
	let filterOpen = $state(false);
	let filterMemberIDs = $state<string[]>([]);
	let filterLabelIDs = $state<string[]>([]);
	const activeFilterCount = $derived(filterMemberIDs.length + filterLabelIDs.length);

	// Day panel
	let selectedDay = $state<Date | null>(null);
	let dayPanelOpen = $state(false);
	let dayQuickTask = $state('');

	// Event form
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

	async function loadData() {
		const from = new Date(viewYear, viewMonth, 1);
		const to = new Date(viewYear, viewMonth + 1, 1);
		// Agenda uses wider window; fetch enough for both
		const agendaFrom = new Date(today); agendaFrom.setDate(agendaFrom.getDate() - 30);
		const agendaTo = new Date(today); agendaTo.setDate(agendaTo.getDate() + 120);
		const fetchFrom = view === 'agenda' ? agendaFrom : from;
		const fetchTo   = view === 'agenda' ? agendaTo   : to;
		try {
			[events, tasks, members, labels] = await Promise.all([
				api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${fetchFrom.toISOString()}&to=${fetchTo.toISOString()}`).then(r => r ?? []),
				api.get<Task[]>(`/api/v1/families/${familyID}/tasks`).then(r => r ?? []),
				api.get<Member[]>(`/api/v1/families/${familyID}/members`).then(r => r ?? []),
				api.get<AppLabel[]>(`/api/v1/families/${familyID}/labels`).then(r => r ?? []),
			]);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load calendar';
		}
	}

	let es: EventSource | null = null;
	onMount(async () => {
		view = window.innerWidth < 768 ? 'agenda' : 'month';
		await loadData();
		if (view === 'agenda') {
			await tick();
			document.getElementById('agenda-today')?.scrollIntoView({ block: 'start' });
		}
		es = new EventSource(sseUrl(`/api/v1/families/${familyID}/stream`));
		es.onmessage = (e) => { if (e.data === 'refresh') loadData(); };
		es.onerror = () => { es?.close(); es = null; };
	});
	onDestroy(() => es?.close());

	async function switchView(v: 'month' | 'agenda') {
		view = v;
		await loadData();
		if (v === 'agenda') {
			await tick();
			document.getElementById('agenda-today')?.scrollIntoView({ block: 'start' });
		}
	}

	function prevMonth() {
		if (viewMonth === 0) { viewYear--; viewMonth = 11; } else { viewMonth--; }
		loadData();
	}
	function nextMonth() {
		if (viewMonth === 11) { viewYear++; viewMonth = 0; } else { viewMonth++; }
		loadData();
	}
	function jumpToToday() {
		viewYear = today.getFullYear();
		viewMonth = today.getMonth();
		if (view === 'agenda') {
			tick().then(() => document.getElementById('agenda-today')?.scrollIntoView({ block: 'start', behavior: 'smooth' }));
		}
	}

	// ── Filtering ────────────────────────────────────────────────────────────
	function filterEvents(evs: CalEvent[]): CalEvent[] {
		return evs.filter(ev => {
			if (filterMemberIDs.length > 0 && !filterMemberIDs.some(id => ev.attendee_ids?.includes(id))) return false;
			if (filterLabelIDs.length > 0 && !filterLabelIDs.some(id => ev.label_ids?.includes(id))) return false;
			return true;
		});
	}
	function filterTasks(tsks: Task[]): Task[] {
		return tsks.filter(t => {
			if (t.status === 'done') return false;
			if (filterMemberIDs.length > 0 && !filterMemberIDs.includes(t.assigned_to ?? '')) return false;
			if (filterLabelIDs.length > 0 && !filterLabelIDs.some(id => t.label_ids?.includes(id))) return false;
			return true;
		});
	}

	// ── Day helpers ──────────────────────────────────────────────────────────
	function dayMs(date: Date) {
		return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
	}
	function eventsForDay(date: Date): CalEvent[] {
		const ms = dayMs(date);
		return filterEvents(events).filter(ev => {
			const s = new Date(ev.start_at); const e = new Date(ev.end_at);
			const sMs = new Date(s.getFullYear(), s.getMonth(), s.getDate()).getTime();
			const eMs = new Date(e.getFullYear(), e.getMonth(), e.getDate()).getTime();
			return sMs <= ms && ms <= eMs;
		}).sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime());
	}
	function tasksForDay(date: Date): Task[] {
		const ms = dayMs(date);
		return filterTasks(tasks).filter(t => {
			if (!t.end_date) return false;
			const [y, m, d] = t.end_date.slice(0, 10).split('-').map(Number);
			return new Date(y, m - 1, d).getTime() === ms;
		});
	}
	function isToday(date: Date) { return dayMs(date) === todayMs; }
	function isPast(date: Date)  { return dayMs(date) < todayMs; }

	// ── Month grid ───────────────────────────────────────────────────────────
	const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	const monthLabel = $derived(new Date(viewYear, viewMonth, 1).toLocaleString('default', { month: 'long', year: 'numeric' }));
	const calDays = $derived((() => {
		const first = new Date(viewYear, viewMonth, 1);
		const last  = new Date(viewYear, viewMonth + 1, 0);
		const days: (Date | null)[] = [];
		for (let i = 0; i < first.getDay(); i++) days.push(null);
		for (let d = 1; d <= last.getDate(); d++) days.push(new Date(viewYear, viewMonth, d));
		while (days.length % 7 !== 0) days.push(null);
		return days;
	})());

	const CAP = 2; // items shown per cell before "+N more"

	// ── Agenda ───────────────────────────────────────────────────────────────
	const agendaDays = $derived((() => {
		const from = new Date(today); from.setDate(from.getDate() - 30);
		const days: Date[] = [];
		for (let i = 0; i < 150; i++) {
			const d = new Date(from); d.setDate(d.getDate() + i);
			days.push(d);
		}
		return days;
	})());

	function fmtAgendaDate(date: Date) {
		return date.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' });
	}
	function fmtTime(iso: string) {
		return new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
	}

	// ── Day panel ────────────────────────────────────────────────────────────
	function selectDay(date: Date) {
		selectedDay = date;
		dayPanelOpen = true;
		dayQuickTask = '';
	}

	async function addTaskForDay() {
		if (!dayQuickTask.trim() || !selectedDay) return;
		const pad = (n: number) => String(n).padStart(2, '0');
		const due = `${selectedDay.getFullYear()}-${pad(selectedDay.getMonth() + 1)}-${pad(selectedDay.getDate())}`;
		try {
			await api.post(`/api/v1/families/${familyID}/tasks`, {
				title: dayQuickTask.trim(), priority: 'medium', label_ids: [], end_date: due,
			});
			dayQuickTask = '';
			loadData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to add task';
		}
	}

	// ── Event form ───────────────────────────────────────────────────────────
	function openNew(date?: Date) {
		editing = null; formTitle = ''; formDesc = ''; formLocation = '';
		formAllDay = false; formLabelIDs = []; startTime = '09:00'; endTime = '10:00';
		if (date) {
			const cd = new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate());
			dateRange = { start: cd, end: cd };
		} else {
			dateRange = { start: undefined, end: undefined };
		}
		dayPanelOpen = false;
		showForm = true;
	}
	function openEdit(ev: CalEvent) {
		editing = ev; formTitle = ev.title; formDesc = ev.description ?? '';
		formLocation = ev.location ?? ''; formAllDay = ev.all_day;
		formLabelIDs = [...(ev.label_ids ?? [])];
		const s = new Date(ev.start_at); const en = new Date(ev.end_at);
		dateRange = { start: toCalDate(ev.start_at), end: toCalDate(ev.end_at) };
		startTime = `${String(s.getHours()).padStart(2, '0')}:${String(s.getMinutes()).padStart(2, '0')}`;
		endTime   = `${String(en.getHours()).padStart(2, '0')}:${String(en.getMinutes()).padStart(2, '0')}`;
		dayPanelOpen = false;
		showForm = true;
	}
	async function saveEvent() {
		if (!formTitle.trim() || !dateRange.start) return;
		saving = true;
		const endDate = dateRange.end ?? dateRange.start;
		try {
			const body = {
				title: formTitle.trim(), description: formDesc, location: formLocation,
				start_at: toISO(dateRange.start, startTime, formAllDay),
				end_at: toISO(endDate, endTime, formAllDay),
				all_day: formAllDay, label_ids: formLabelIDs,
			};
			if (editing) {
				await api.patch(`/api/v1/families/${familyID}/events/${editing.id}`, body);
			} else {
				await api.post(`/api/v1/families/${familyID}/events`, body);
			}
			showForm = false;
			loadData();
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
			loadData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete event';
		}
	}
</script>

{#if error}
	<p class="text-sm text-destructive mb-2">{error}</p>
{/if}

<!-- Header -->
<div class="flex items-center justify-between mb-3 gap-2">
	<div class="flex items-center gap-1.5">
		<!-- View toggle -->
		<div class="flex rounded-md border border-border overflow-hidden text-sm shrink-0">
			<button
				onclick={() => switchView('month')}
				class="px-3 py-1.5 transition-colors cursor-pointer {view === 'month' ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted'}"
			>Month</button>
			<button
				onclick={() => switchView('agenda')}
				class="px-3 py-1.5 transition-colors cursor-pointer {view === 'agenda' ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted'}"
			>Agenda</button>
		</div>

		{#if view === 'month'}
			<Button variant="outline" size="sm" onclick={prevMonth}>‹</Button>
			<span class="text-sm font-medium w-32 text-center hidden sm:block">{monthLabel}</span>
			<Button variant="outline" size="sm" onclick={nextMonth}>›</Button>
		{/if}

		<Button variant="outline" size="sm" onclick={jumpToToday}>Today</Button>
	</div>

	<div class="flex items-center gap-1.5 shrink-0">
		<!-- Filter -->
		<Popover.Root bind:open={filterOpen}>
			<Popover.Trigger>
				<Button variant="outline" size="sm" class="gap-1.5 relative">
					<Filter class="w-3.5 h-3.5" />
					<span class="hidden sm:inline">Filter</span>
					{#if activeFilterCount > 0}
						<span class="absolute -top-1.5 -right-1.5 w-4 h-4 rounded-full bg-primary text-primary-foreground text-[10px] flex items-center justify-center font-bold">{activeFilterCount}</span>
					{/if}
				</Button>
			</Popover.Trigger>
			<Popover.Content class="w-64 p-4 flex flex-col gap-4" align="end">
				{#if members.filter(m => !m.virtual).length > 1}
					<div class="flex flex-col gap-2">
						<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Who</p>
						{#each members as m (m.user_id)}
							<label class="flex items-center gap-2 text-sm cursor-pointer">
								<Checkbox
									checked={filterMemberIDs.includes(m.user_id)}
									onCheckedChange={(v) => {
										filterMemberIDs = v
											? [...filterMemberIDs, m.user_id]
											: filterMemberIDs.filter(id => id !== m.user_id);
									}}
								/>
								{m.name}
							</label>
						{/each}
					</div>
				{/if}
				{#if labels.length > 0}
					<div class="flex flex-col gap-2">
						<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Labels</p>
						{#each labels as lbl (lbl.id)}
							<label class="flex items-center gap-2 text-sm cursor-pointer">
								<Checkbox
									checked={filterLabelIDs.includes(lbl.id)}
									onCheckedChange={(v) => {
										filterLabelIDs = v
											? [...filterLabelIDs, lbl.id]
											: filterLabelIDs.filter(id => id !== lbl.id);
									}}
								/>
								<span class="w-2.5 h-2.5 rounded-full {dotClass(lbl.color)} shrink-0"></span>
								{lbl.name}
							</label>
						{/each}
					</div>
				{/if}
				{#if activeFilterCount > 0}
					<button
						onclick={() => { filterMemberIDs = []; filterLabelIDs = []; }}
						class="text-xs text-muted-foreground hover:text-foreground transition-colors text-left cursor-pointer"
					>Clear all</button>
				{/if}
			</Popover.Content>
		</Popover.Root>

		<Button size="sm" onclick={() => openNew()}>+ Event</Button>
	</div>
</div>

{#if view === 'month'}
	<!-- Month label (mobile) -->
	<p class="text-sm font-medium text-center mb-2 sm:hidden">{monthLabel}</p>

	<!-- Month grid -->
	<div class="grid grid-cols-7 border-l border-t rounded-lg overflow-hidden text-xs">
		{#each DAYS as d}
			<div class="border-r border-b px-1 py-1.5 text-center text-muted-foreground font-semibold bg-muted/40 text-[11px] uppercase tracking-wide">{d}</div>
		{/each}
		{#each calDays as day}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<div
				class="border-r border-b min-h-24 flex flex-col cursor-pointer hover:bg-muted/20 transition-colors group"
				onclick={() => day && selectDay(day)}
				role="button"
				tabindex="0"
				onkeydown={(e) => e.key === 'Enter' && day && selectDay(day)}
			>
				{#if day}
					<div class="px-1.5 pt-1.5 pb-1">
						<span class="text-xs font-medium inline-flex items-center justify-center w-6 h-6 rounded-full
							{isToday(day) ? 'bg-primary text-primary-foreground font-bold' : isPast(day) ? 'text-muted-foreground' : 'text-foreground'}">
							{day.getDate()}
						</span>
					</div>
					{@const dayEvs = eventsForDay(day)}
					{@const dayTsks = tasksForDay(day)}
					{@const allItems = [...dayEvs.map(e => ({ type: 'event' as const, item: e })), ...dayTsks.map(t => ({ type: 'task' as const, item: t }))]}
					{@const visible = allItems.slice(0, CAP)}
					{@const overflow = allItems.length - CAP}
					<div class="flex flex-col gap-0.5 px-1 pb-1 flex-1">
						{#each visible as { type, item }}
							{#if type === 'event'}
								<button
									class="w-full text-left truncate rounded px-1.5 py-0.5 bg-blue-500/15 text-blue-700 dark:text-blue-300 text-[11px] font-medium leading-5 hover:bg-blue-500/25 transition-colors"
									onclick={(e) => { e.stopPropagation(); openEdit(item as CalEvent); }}
								>{(item as CalEvent).title}</button>
							{:else}
								<div class="flex items-center gap-0.5 rounded px-1 py-0.5 bg-primary/10 text-primary text-[11px] font-medium leading-5 truncate">
									<SquareCheckBig class="w-2.5 h-2.5 shrink-0" />
									<span class="truncate">{(item as Task).title}</span>
								</div>
							{/if}
						{/each}
						{#if overflow > 0}
							<button
								class="text-[11px] text-muted-foreground hover:text-foreground transition-colors text-left px-1 cursor-pointer"
								onclick={(e) => { e.stopPropagation(); day && selectDay(day); }}
							>+{overflow} more</button>
						{/if}
					</div>
				{/if}
			</div>
		{/each}
	</div>

{:else}
	<!-- Agenda view -->
	<div class="flex flex-col">
		{#each agendaDays as day (day.toDateString())}
			{@const dayEvs = eventsForDay(day)}
			{@const dayTsks = tasksForDay(day)}
			{@const hasItems = dayEvs.length > 0 || dayTsks.length > 0}
			<div id={isToday(day) ? 'agenda-today' : undefined} class="flex gap-3 py-3 border-b border-border last:border-0">
				<!-- Date column -->
				<div class="w-16 shrink-0 pt-0.5">
					<p class="text-xs font-semibold {isToday(day) ? 'text-primary' : isPast(day) ? 'text-muted-foreground/50' : 'text-muted-foreground'}">
						{day.toLocaleDateString(undefined, { weekday: 'short' }).toUpperCase()}
					</p>
					<p class="text-lg font-bold leading-tight {isToday(day) ? 'text-primary' : isPast(day) ? 'text-muted-foreground/50' : 'text-foreground'}">
						{day.getDate()}
					</p>
					<p class="text-[10px] {isPast(day) ? 'text-muted-foreground/50' : 'text-muted-foreground'}">
						{day.toLocaleDateString(undefined, { month: 'short' })}
					</p>
				</div>

				<!-- Items column -->
				<div class="flex-1 flex flex-col gap-1.5 min-w-0">
					{#if !hasItems}
						<p class="text-xs text-muted-foreground/40 pt-1.5">Nothing scheduled</p>
					{/if}
					{#each dayEvs as ev (ev.id)}
						<button
							onclick={() => openEdit(ev)}
							class="w-full text-left flex items-start gap-2 rounded-lg border-l-4 border-l-blue-400 bg-card border border-border px-3 py-2 hover:bg-muted/40 transition-colors cursor-pointer"
						>
							<div class="flex-1 min-w-0">
								<p class="text-sm font-medium truncate">{ev.title}</p>
								{#if !ev.all_day}
									<p class="text-xs text-muted-foreground">{fmtTime(ev.start_at)} – {fmtTime(ev.end_at)}</p>
								{/if}
							</div>
						</button>
					{/each}
					{#each dayTsks as task (task.id)}
						<div class="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-2">
							<SquareCheckBig class="w-4 h-4 text-muted-foreground shrink-0" />
							<p class="text-sm flex-1 truncate">{task.title}</p>
						</div>
					{/each}
				</div>

				<!-- Add event shortcut on today -->
				{#if isToday(day)}
					<button
						onclick={() => openNew(day)}
						class="shrink-0 text-xs text-muted-foreground hover:text-foreground transition-colors pt-1.5 cursor-pointer"
						aria-label="Add event today"
					>+</button>
				{/if}
			</div>
		{/each}
	</div>
{/if}

<!-- Day panel -->
<Dialog.Root bind:open={dayPanelOpen}>
	<Dialog.Portal>
		<Dialog.Overlay />
		<Dialog.Content class="sm:max-w-sm">
			{#if selectedDay}
				<Dialog.Header>
					<Dialog.Title>{fmtAgendaDate(selectedDay)}</Dialog.Title>
				</Dialog.Header>
				{@const dayEvs = eventsForDay(selectedDay)}
				{@const dayTsks = tasksForDay(selectedDay)}
				<div class="flex flex-col gap-2 py-2 max-h-72 overflow-y-auto">
					{#if dayEvs.length === 0 && dayTsks.length === 0}
						<p class="text-sm text-muted-foreground py-4 text-center">Nothing scheduled.</p>
					{/if}
					{#each dayEvs as ev (ev.id)}
						<button
							onclick={() => openEdit(ev)}
							class="w-full text-left flex items-start gap-2 rounded-lg border-l-4 border-l-blue-400 bg-muted/30 px-3 py-2 hover:bg-muted/50 transition-colors cursor-pointer"
						>
							<div class="flex-1 min-w-0">
								<p class="text-sm font-medium truncate">{ev.title}</p>
								{#if !ev.all_day}
									<p class="text-xs text-muted-foreground">{fmtTime(ev.start_at)} – {fmtTime(ev.end_at)}</p>
								{/if}
							</div>
						</button>
					{/each}
					{#each dayTsks as task (task.id)}
						<div class="flex items-center gap-2 rounded-lg bg-muted/30 px-3 py-2">
							<SquareCheckBig class="w-4 h-4 text-muted-foreground shrink-0" />
							<p class="text-sm flex-1 truncate">{task.title}</p>
						</div>
					{/each}
				</div>
				<form onsubmit={(e) => { e.preventDefault(); addTaskForDay(); }} class="mt-2">
					<Input bind:value={dayQuickTask} placeholder="Add a task for this day…" class="bg-muted/20 border-dashed focus-visible:border-solid" />
				</form>
				<Dialog.Footer class="mt-3">
					<Button variant="outline" size="sm" onclick={() => { if (selectedDay) openNew(selectedDay); }}>
						+ Event
					</Button>
				</Dialog.Footer>
			{/if}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

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
							<RangeCalendar bind:value={dateRange} onValueChange={() => { if (dateRange.start && dateRange.end) pickerOpen = false; }} />
						</Popover.Content>
					</Popover.Root>
				</div>
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
											? formLabelIDs.filter(id => id !== lbl.id)
											: [...formLabelIDs, lbl.id];
									}}
									class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-all border cursor-pointer
										{formLabelIDs.includes(lbl.id) ? 'border-foreground ring-1 ring-foreground ' + chipClass(lbl.color) : 'border-transparent opacity-50 ' + chipClass(lbl.color)}"
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
