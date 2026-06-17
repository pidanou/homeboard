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

	// ── View ─────────────────────────────────────────────────────────────────
	let view = $state<'month' | 'week' | 'day' | 'agenda'>('month');

	// ── Month nav ─────────────────────────────────────────────────────────────
	let viewYear = $state(today.getFullYear());
	let viewMonth = $state(today.getMonth());

	// ── Week / day nav ────────────────────────────────────────────────────────
	function getWeekStart(d: Date): Date {
		const s = new Date(d.getFullYear(), d.getMonth(), d.getDate());
		s.setDate(s.getDate() - s.getDay()); // Sunday
		return s;
	}
	let weekStart = $state(getWeekStart(today));
	let dayViewDate = $state(new Date(today.getFullYear(), today.getMonth(), today.getDate()));

	const viewDays = $derived(
		view === 'week'
			? Array.from({ length: 7 }, (_, i) => { const d = new Date(weekStart); d.setDate(d.getDate() + i); return d; })
			: [dayViewDate]
	);

	// ── Time grid ─────────────────────────────────────────────────────────────
	const HOUR_H = 56; // px per hour
	const HOURS = Array.from({ length: 24 }, (_, i) => i);
	let timeGridEl: HTMLElement | null = null;
	let now = $state(new Date());

	function fmtHour(h: number): string {
		if (h === 0) return '12am';
		if (h === 12) return '12pm';
		return h < 12 ? `${h}am` : `${h - 12}pm`;
	}
	function eventTop(ev: CalEvent): number {
		const d = new Date(ev.start_at);
		return (d.getHours() + d.getMinutes() / 60) * HOUR_H;
	}
	function eventHeight(ev: CalEvent): number {
		const mins = (new Date(ev.end_at).getTime() - new Date(ev.start_at).getTime()) / 60000;
		return Math.max(mins, 30) * (HOUR_H / 60);
	}
	function scrollToCurrentTime() {
		const target = Math.max((now.getHours() - 1) * HOUR_H, 0);
		tick().then(() => timeGridEl?.scrollTo({ top: target }));
	}

	// ── Data ──────────────────────────────────────────────────────────────────
	let events = $state<CalEvent[]>([]);
	let tasks = $state<Task[]>([]);
	let members = $state<Member[]>([]);
	let labels = $state<AppLabel[]>([]);
	let error = $state('');

	// ── Filters ───────────────────────────────────────────────────────────────
	let filterOpen = $state(false);
	let filterMemberIDs = $state<string[]>([]);
	let filterLabelIDs = $state<string[]>([]);
	const activeFilterCount = $derived(filterMemberIDs.length + filterLabelIDs.length);

	// ── Day panel ─────────────────────────────────────────────────────────────
	let selectedDay = $state<Date | null>(null);
	let dayPanelOpen = $state(false);
	let dayQuickTask = $state('');

	// ── Event form ────────────────────────────────────────────────────────────
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

	// ── Period label ──────────────────────────────────────────────────────────
	const monthLabel = $derived(new Date(viewYear, viewMonth, 1).toLocaleString('default', { month: 'long', year: 'numeric' }));
	const periodLabel = $derived((() => {
		if (view === 'month') return monthLabel;
		if (view === 'week') {
			const end = new Date(weekStart); end.setDate(end.getDate() + 6);
			const s = weekStart.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
			const e = end.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
			return `${s} – ${e}`;
		}
		if (view === 'day') return dayViewDate.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' });
		return '';
	})());

	// ── Load data ─────────────────────────────────────────────────────────────
	async function loadData() {
		let fetchFrom: Date, fetchTo: Date;
		if (view === 'month') {
			fetchFrom = new Date(viewYear, viewMonth, 1);
			fetchTo   = new Date(viewYear, viewMonth + 1, 1);
		} else if (view === 'week') {
			fetchFrom = new Date(weekStart);
			fetchTo   = new Date(weekStart); fetchTo.setDate(fetchTo.getDate() + 7);
		} else if (view === 'day') {
			fetchFrom = new Date(dayViewDate); fetchFrom.setHours(0, 0, 0, 0);
			fetchTo   = new Date(dayViewDate); fetchTo.setDate(fetchTo.getDate() + 1);
		} else {
			fetchFrom = new Date(today); fetchFrom.setDate(fetchFrom.getDate() - 30);
			fetchTo   = new Date(today); fetchTo.setDate(fetchTo.getDate() + 120);
		}
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

	// ── SSE ───────────────────────────────────────────────────────────────────
	let es: EventSource | null = null;
	let clockInterval: ReturnType<typeof setInterval>;

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
		clockInterval = setInterval(() => { now = new Date(); }, 60000);
	});
	onDestroy(() => { es?.close(); clearInterval(clockInterval); });

	// ── Navigation ────────────────────────────────────────────────────────────
	async function switchView(v: typeof view) {
		view = v;
		await loadData();
		if (v === 'agenda') {
			await tick();
			document.getElementById('agenda-today')?.scrollIntoView({ block: 'start' });
		} else if (v === 'week' || v === 'day') {
			scrollToCurrentTime();
		}
	}

	function prevPeriod() {
		if (view === 'month') {
			if (viewMonth === 0) { viewYear--; viewMonth = 11; } else { viewMonth--; }
			loadData();
		} else if (view === 'week') {
			weekStart = new Date(weekStart); weekStart.setDate(weekStart.getDate() - 7);
			loadData();
		} else if (view === 'day') {
			dayViewDate = new Date(dayViewDate); dayViewDate.setDate(dayViewDate.getDate() - 1);
			loadData();
		}
	}
	function nextPeriod() {
		if (view === 'month') {
			if (viewMonth === 11) { viewYear++; viewMonth = 0; } else { viewMonth++; }
			loadData();
		} else if (view === 'week') {
			weekStart = new Date(weekStart); weekStart.setDate(weekStart.getDate() + 7);
			loadData();
		} else if (view === 'day') {
			dayViewDate = new Date(dayViewDate); dayViewDate.setDate(dayViewDate.getDate() + 1);
			loadData();
		}
	}

	function jumpToToday() {
		viewYear = today.getFullYear(); viewMonth = today.getMonth();
		weekStart = getWeekStart(today);
		dayViewDate = new Date(today.getFullYear(), today.getMonth(), today.getDate());
		if (view === 'agenda') {
			tick().then(() => document.getElementById('agenda-today')?.scrollIntoView({ block: 'start', behavior: 'smooth' }));
		} else if (view === 'week' || view === 'day') {
			loadData().then(() => scrollToCurrentTime());
		}
	}

	// ── Swipe ─────────────────────────────────────────────────────────────────
	let touchStartX = 0;
	function onTouchStart(e: TouchEvent) { touchStartX = e.touches[0].clientX; }
	function onTouchEnd(e: TouchEvent) {
		const dx = e.changedTouches[0].clientX - touchStartX;
		if (Math.abs(dx) > 60) { dx > 0 ? prevPeriod() : nextPeriod(); }
	}

	// ── Filtering ─────────────────────────────────────────────────────────────
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

	// ── Day helpers ───────────────────────────────────────────────────────────
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

	// ── Month grid ────────────────────────────────────────────────────────────
	const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	const calDays = $derived((() => {
		const first = new Date(viewYear, viewMonth, 1);
		const last  = new Date(viewYear, viewMonth + 1, 0);
		const days: (Date | null)[] = [];
		for (let i = 0; i < first.getDay(); i++) days.push(null);
		for (let d = 1; d <= last.getDate(); d++) days.push(new Date(viewYear, viewMonth, d));
		while (days.length % 7 !== 0) days.push(null);
		return days;
	})());

	const CAP = 2;

	// ── Week/day all-day row visibility ───────────────────────────────────────
	const hasAllDay = $derived(
		viewDays.some(d => eventsForDay(d).some(ev => ev.all_day) || tasksForDay(d).length > 0)
	);

	// ── Agenda ────────────────────────────────────────────────────────────────
	const agendaDays = $derived((() => {
		const from = new Date(today); from.setDate(from.getDate() - 30);
		return Array.from({ length: 150 }, (_, i) => { const d = new Date(from); d.setDate(d.getDate() + i); return d; });
	})());

	function fmtAgendaDate(date: Date) {
		return date.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' });
	}
	function fmtTime(iso: string) {
		return new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
	}

	// ── Day panel ─────────────────────────────────────────────────────────────
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

	// ── Time grid click ───────────────────────────────────────────────────────
	function handleTimeGridClick(e: MouseEvent & { currentTarget: HTMLElement }, day: Date) {
		const rect = e.currentTarget.getBoundingClientRect();
		const totalMins = Math.round((e.clientY - rect.top) / HOUR_H * 60 / 15) * 15;
		const h = Math.min(Math.floor(totalMins / 60), 23);
		const m = totalMins % 60;
		const pad = (n: number) => String(n).padStart(2, '0');
		openNew(day, { start: `${pad(h)}:${pad(m)}`, end: `${pad(Math.min(h + 1, 23))}:${pad(m)}` });
	}

	// ── Event form ────────────────────────────────────────────────────────────
	function openNew(date?: Date, time?: { start: string; end: string }) {
		editing = null; formTitle = ''; formDesc = ''; formLocation = '';
		formAllDay = false; formLabelIDs = [];
		startTime = time?.start ?? '09:00';
		endTime   = time?.end   ?? '10:00';
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
<div class="flex items-center justify-between mb-3 gap-2 flex-wrap">
	<div class="flex items-center gap-1.5 flex-wrap">
		<!-- View toggle -->
		<div class="flex rounded-md border border-border overflow-hidden text-sm shrink-0">
			{#each [['month','M','Month'],['week','W','Week'],['day','D','Day'],['agenda','A','Agenda']] as [v, short, long]}
				<button
					onclick={() => switchView(v as typeof view)}
					class="px-2.5 py-1.5 transition-colors cursor-pointer {view === v ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted'}"
				>
					<span class="sm:hidden">{short}</span>
					<span class="hidden sm:inline">{long}</span>
				</button>
			{/each}
		</div>

		{#if view !== 'agenda'}
			<Button variant="outline" size="sm" onclick={prevPeriod}>‹</Button>
			<span class="text-sm font-medium hidden sm:block max-w-48 truncate text-center">{periodLabel}</span>
			<Button variant="outline" size="sm" onclick={nextPeriod}>›</Button>
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

<!-- Period label (mobile, non-agenda) -->
{#if view !== 'agenda'}
	<p class="text-sm font-medium text-center mb-2 sm:hidden truncate">{periodLabel}</p>
{/if}

{#if view === 'month'}
	<!-- ── Month grid ──────────────────────────────────────────────────────── -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="grid grid-cols-7 border-l border-t rounded-lg overflow-hidden text-xs"
		ontouchstart={onTouchStart}
		ontouchend={onTouchEnd}
	>
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

{:else if view === 'week' || view === 'day'}
	<!-- ── Week / Day time grid ────────────────────────────────────────────── -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="overflow-auto rounded-lg border border-border"
		style="max-height: calc(100vh - 12rem)"
		bind:this={timeGridEl}
		ontouchstart={onTouchStart}
		ontouchend={onTouchEnd}
	>
		<!-- Day column headers -->
		<div class="sticky top-0 z-20 bg-background border-b flex" style="padding-left: 3rem">
			{#each viewDays as day (day.toDateString())}
				<button
					class="flex-1 text-center py-2 border-l first:border-l-0 cursor-pointer hover:bg-muted/30 transition-colors"
					onclick={() => { dayViewDate = new Date(day.getFullYear(), day.getMonth(), day.getDate()); switchView('day'); }}
				>
					<p class="text-[10px] uppercase text-muted-foreground">{day.toLocaleDateString(undefined, { weekday: 'short' })}</p>
					<p class="text-lg font-bold leading-none mt-0.5 {isToday(day) ? 'text-primary' : isPast(day) ? 'text-muted-foreground' : 'text-foreground'}">{day.getDate()}</p>
				</button>
			{/each}
		</div>

		<!-- All-day + tasks row (shown when non-empty) -->
		{#if hasAllDay}
			<div class="flex border-b" style="padding-left: 3rem">
				{#each viewDays as day (day.toDateString())}
					<div class="flex-1 p-1 border-l first:border-l-0 flex flex-col gap-0.5 min-h-8">
						{#each eventsForDay(day).filter(ev => ev.all_day) as ev (ev.id)}
							<button
								onclick={() => openEdit(ev)}
								class="w-full text-left truncate rounded px-1.5 py-0.5 bg-blue-500/15 text-blue-700 dark:text-blue-300 text-[11px] font-medium cursor-pointer hover:bg-blue-500/25"
							>{ev.title}</button>
						{/each}
						{#each tasksForDay(day) as task (task.id)}
							<div class="flex items-center gap-0.5 rounded px-1 py-0.5 bg-primary/10 text-primary text-[11px] truncate">
								<SquareCheckBig class="w-2.5 h-2.5 shrink-0" /><span class="truncate">{task.title}</span>
							</div>
						{/each}
					</div>
				{/each}
			</div>
		{/if}

		<!-- Time grid -->
		<div class="flex">
			<!-- Hour labels -->
			<div class="w-12 shrink-0 relative select-none" style="height: {24 * HOUR_H}px">
				{#each HOURS as h}
					<div
						class="absolute text-[10px] text-muted-foreground text-right pr-2 leading-none w-full"
						style="top: {h * HOUR_H - 6}px"
					>{h > 0 ? fmtHour(h) : ''}</div>
				{/each}
			</div>

			<!-- Day columns -->
			{#each viewDays as day (day.toDateString())}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<div
					class="flex-1 relative border-l first:border-l-0"
					style="height: {24 * HOUR_H}px; min-width: 0"
					role="button"
					tabindex="0"
					onclick={(e) => handleTimeGridClick(e, day)}
				>
					<!-- Hour lines -->
					{#each HOURS as h}
						<div class="absolute w-full border-t border-border/40 pointer-events-none" style="top: {h * HOUR_H}px" />
						<div class="absolute w-full border-t border-dashed border-border/20 pointer-events-none" style="top: {h * HOUR_H + HOUR_H / 2}px" />
					{/each}

					<!-- Current time indicator -->
					{#if isToday(day)}
						{@const nowFrac = now.getHours() + now.getMinutes() / 60}
						<div class="absolute w-full z-10 flex items-center pointer-events-none" style="top: {nowFrac * HOUR_H}px">
							<div class="w-2 h-2 rounded-full bg-red-500 shrink-0 -ml-1" />
							<div class="flex-1 h-0.5 bg-red-500" />
						</div>
					{/if}

					<!-- Timed events -->
					{#each eventsForDay(day).filter(ev => !ev.all_day) as ev (ev.id)}
						<button
							onclick={(e) => { e.stopPropagation(); openEdit(ev); }}
							class="absolute left-0.5 right-0.5 rounded-sm bg-blue-500/20 border-l-2 border-blue-500 px-1 py-0.5 text-left overflow-hidden hover:bg-blue-500/30 transition-colors cursor-pointer"
							style="top: {eventTop(ev)}px; height: {eventHeight(ev)}px; z-index: 5; min-height: 20px"
						>
							<p class="text-[11px] font-medium text-blue-700 dark:text-blue-300 leading-tight truncate">{ev.title}</p>
							{#if eventHeight(ev) > 32}
								<p class="text-[10px] text-blue-600/70 dark:text-blue-400/70">{fmtTime(ev.start_at)}</p>
							{/if}
						</button>
					{/each}
				</div>
			{/each}
		</div>
	</div>

{:else}
	<!-- ── Agenda view ─────────────────────────────────────────────────────── -->
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
				<Dialog.Footer class="mt-3 flex gap-2 justify-between sm:justify-between">
					<Button variant="ghost" size="sm" onclick={() => {
						if (selectedDay) {
							dayViewDate = new Date(selectedDay.getFullYear(), selectedDay.getMonth(), selectedDay.getDate());
							switchView('day');
						}
						dayPanelOpen = false;
					}}>Day view</Button>
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
