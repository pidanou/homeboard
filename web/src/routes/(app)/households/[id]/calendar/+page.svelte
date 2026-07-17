<script lang="ts">
	import '@event-calendar/core/index.css';
	import { page } from '$app/stores';
	import { onMount, onDestroy, tick } from 'svelte';
	import { Calendar, DayGrid, TimeGrid, Interaction } from '@event-calendar/core';
	import { api, sseUrl } from '$lib/api/client';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import * as Popover from '$lib/components/ui/popover';
	import { Calendar as DatePicker } from '$lib/components/ui/calendar';
	import { CalendarDate, type DateValue } from '@internationalized/date';
	import { X, CalendarDays, ChevronDown, Pencil, Trash2 } from 'lucide-svelte';
	import type { CalEvent, Task, Member, AppCategory } from '$lib/types';
	import { dotClass, CATEGORY_HEX } from '$lib/categories';
	import { fmtTime, taskHasTime, fmtWeekdayDate, hour12Option } from '$lib/dates';
	import EditDialog from '$lib/components/EditDialog.svelte';
	import CreateDialog from '$lib/components/CreateDialog.svelte';
	import EventCard from '$lib/components/EventCard.svelte';
	import TaskCard from '$lib/components/TaskCard.svelte';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import * as msg from '$lib/paraglide/messages.js';

	function categoryHex(categoryID: string | undefined): string | null {
		if (!categoryID) return null;
		const cat = categories.find(c => c.id === categoryID);
		return cat ? (CATEGORY_HEX[cat.color] ?? null) : null;
	}

	const familyID = $derived($page.params.id ?? '');
	const today = new Date();
	const todayMs = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime();

	// ── View ──────────────────────────────────────────────────────────────────
	type AppView = 'month' | 'week' | 'day' | 'agenda';
	const EC_VIEW: Record<AppView, string> = {
		month: 'dayGridMonth', week: 'timeGridWeek', day: 'timeGridDay', agenda: '',
	};
	let appView = $state<AppView>('month');
	let periodLabel = $state('');

	// ── Data ──────────────────────────────────────────────────────────────────
	let events = $state<CalEvent[]>([]);
	let tasks = $state<Task[]>([]);
	let members = $state<Member[]>([]);
	let categories = $state<AppCategory[]>([]);

	// Current visible date range (set by EC's datesSet callback)
	let viewStart    = $state(new Date(today.getFullYear(), today.getMonth(), 1));
	let viewEnd      = $state(new Date(today.getFullYear(), today.getMonth() + 1, 1));
	let currentStart = $state(new Date(today.getFullYear(), today.getMonth(), 1));

	// ── Filters ───────────────────────────────────────────────────────────────
	let filterTypes = $state(new Set<'task' | 'event' | 'birthday'>());
	let showCompleted = $state(false);
	let filterMemberIDs = $state<string[]>([]);
	let filterCategoryIDs = $state(new Set<string>());
	const someFilterActive = $derived(filterTypes.size > 0 || filterMemberIDs.length > 0 || filterCategoryIDs.size > 0);

	function toggleType(t: 'task' | 'event' | 'birthday') {
		const next = new Set(filterTypes);
		next.has(t) ? next.delete(t) : next.add(t);
		filterTypes = next;
	}
	function toggleMember(id: string) {
		filterMemberIDs = filterMemberIDs.includes(id) ? filterMemberIDs.filter(x => x !== id) : [...filterMemberIDs, id];
	}
	function toggleCategory(id: string) {
		const next = new Set(filterCategoryIDs);
		next.has(id) ? next.delete(id) : next.add(id);
		filterCategoryIDs = next;
	}
	function clearFilters() { filterTypes = new Set(); filterMemberIDs = []; filterCategoryIDs = new Set(); showCompleted = false; }
	function chipCls(active: boolean) {
		if (active) return 'ring-1 ring-foreground opacity-100';
		return someFilterActive ? 'opacity-30' : 'opacity-70 hover:opacity-100';
	}
	function initials(name: string) { return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase(); }

	// ── Agenda ────────────────────────────────────────────────────────────────
	// Past: loaded on demand via "Load previous month" button (explicit, no scroll tricks)
	// Future: loaded automatically via IntersectionObserver on a bottom sentinel
	let agendaStart = $state(new Date(today.getFullYear(), today.getMonth(), today.getDate()));
	let agendaEnd   = $state(new Date(today.getFullYear(), today.getMonth() + 1, today.getDate()));
	let agendaEvents = $state<CalEvent[]>([]);
	let agendaBottomSentinel = $state<HTMLElement | null>(null);
	let agendaLoadingTop    = $state(false);
	let agendaLoadingBottom = $state(false);
	let agendaReady = $state(false);

	$effect(() => {
		if (!agendaBottomSentinel || appView !== 'agenda') return;
		const io = new IntersectionObserver((entries) => {
			if (!entries[0].isIntersecting || agendaLoadingBottom) return;
			extendAgendaForward();
		}, { rootMargin: '300px' });
		io.observe(agendaBottomSentinel);
		return () => io.disconnect();
	});

	async function loadAgenda() {
		agendaLoadingBottom = true;
		try {
			const [evs, tsks, mems, cats] = await Promise.all([
				api.get<CalEvent[]>(`/api/v1/households/${familyID}/events?from=${agendaStart.toISOString()}&to=${agendaEnd.toISOString()}`).then(r => r ?? []),
				api.get<Task[]>(`/api/v1/households/${familyID}/tasks`).then(r => r ?? []),
				members.length ? Promise.resolve(members) : api.get<Member[]>(`/api/v1/households/${familyID}/members`).then(r => r ?? []),
				api.get<AppCategory[]>(`/api/v1/households/${familyID}/categories`).then(r => r ?? []),
			]);
			agendaEvents = evs; tasks = tsks; members = mems; categories = cats;
			agendaReady = true;
		} finally {
			agendaLoadingBottom = false;
		}
	}

	async function extendAgendaForward() {
		if (agendaLoadingBottom) return;
		agendaLoadingBottom = true;
		const newEnd = new Date(agendaEnd);
		newEnd.setMonth(newEnd.getMonth() + 1);
		try {
			const evs = await api.get<CalEvent[]>(`/api/v1/households/${familyID}/events?from=${agendaEnd.toISOString()}&to=${newEnd.toISOString()}`).then(r => r ?? []);
			agendaEvents = [...agendaEvents, ...evs];
			agendaEnd = newEnd;
		} finally {
			agendaLoadingBottom = false;
		}
	}

	// Explicit button — no scroll detection needed. Scroll preservation is reliable
	// because the user is stationary (they clicked a button, not scrolling).
	async function extendAgendaBack() {
		if (agendaLoadingTop) return;
		agendaLoadingTop = true;
		const newStart = new Date(agendaStart);
		newStart.setMonth(newStart.getMonth() - 1);
		try {
			const scrollEl = contentEl;
			const prevHeight = scrollEl?.scrollHeight ?? 0;
			const prevTop    = scrollEl?.scrollTop ?? 0;
			const evs = await api.get<CalEvent[]>(`/api/v1/households/${familyID}/events?from=${newStart.toISOString()}&to=${agendaStart.toISOString()}`).then(r => r ?? []);
			agendaEvents = [...evs, ...agendaEvents];
			agendaStart = newStart;
			await tick();
			if (scrollEl) scrollEl.scrollTop = prevTop + (scrollEl.scrollHeight - prevHeight);
		} finally {
			agendaLoadingTop = false;
		}
	}

	function scrollToToday() {
		const el = document.querySelector('[data-agenda-today]') as HTMLElement | null;
		const scrollEl = contentEl;
		if (!el || !scrollEl) return;
		scrollEl.scrollTo({ top: scrollEl.scrollTop + el.getBoundingClientRect().top - scrollEl.getBoundingClientRect().top - 16, behavior: 'smooth' });
	}

	type AgendaGroup = { dayMs: number; label: string; events: CalEvent[]; tasks: Task[] };

	const agendaGroups = $derived((() => {
		if (appView !== 'agenda') return [] as AgendaGroup[];
		const showEvents = filterTypes.size === 0 || filterTypes.has('event');
		const showTasks  = filterTypes.size === 0 || filterTypes.has('task');
		const showBirthdays = filterTypes.size === 0 || filterTypes.has('birthday');
		const byMemberEv = (ev: CalEvent) =>
			filterMemberIDs.length === 0 || (ev.attendee_ids ?? []).some(id => filterMemberIDs.includes(id));
		const byMember = (id: string | undefined) =>
			filterMemberIDs.length === 0 || (!!id && filterMemberIDs.includes(id));
		const byCat = (id: string | undefined) =>
			filterCategoryIDs.size === 0 || (!!id && filterCategoryIDs.has(id));
		const startMs = agendaStart.getTime();
		const endMs   = agendaEnd.getTime();

		// Pre-fill every day in the loaded range — past days appear when explicitly loaded via button
		const dayMap = new Map<number, { evs: CalEvent[]; tsks: Task[] }>();
		for (let ms = startMs; ms <= endMs; ms += 86400000) {
			dayMap.set(ms, { evs: [], tsks: [] });
		}

		if (showEvents || showBirthdays) {
			for (const ev of agendaEvents) {
				if (ev.birthday_of ? !showBirthdays : !showEvents) continue;
				if (!byMemberEv(ev) || !byCat(ev.category_id)) continue;
				const d = new Date(ev.start_at);
				const dayMs = new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
				dayMap.get(dayMs)?.evs.push(ev);
			}
		}
		if (showTasks) {
			for (const t of tasks) {
				if (!t.end_date) continue;
				if (t.status === 'done' && !showCompleted) continue;
				if (!byMember(t.assigned_to) || !byCat(t.category_id)) continue;
				const d = new Date(t.end_date);
				const dayMs = new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
				dayMap.get(dayMs)?.tsks.push(t);
			}
		}

		return [...dayMap.entries()]
			.sort(([a], [b]) => a - b)
			.map(([dayMs, { evs, tsks }]) => ({
				dayMs,
				label: fmtWeekdayDate(new Date(dayMs)),
				events: evs.sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime()),
				tasks: tsks,
			}));
	})());

	// ── Dialogs ───────────────────────────────────────────────────────────────
	let editDialog: {
		openTask: (t: Task) => void; openEvent: (e: CalEvent) => void;
		deleteTask: (t: Task) => void; deleteEvent: (e: CalEvent) => void;
	} | undefined = $state();
	let createDialog: { open: (t?: 'task' | 'event', start?: Date, end?: Date, allDay?: boolean) => void } | undefined = $state();

	let previewOpen = $state(false);
	let previewAnchor = $state<{ getBoundingClientRect: () => DOMRect } | null>(null);
	let previewItem = $state<{ kind: 'task'; data: Task } | { kind: 'event'; data: CalEvent } | null>(null);

	// Dismissing the preview via an outside click on the calendar itself would
	// otherwise let that same pointerdown reach the calendar's own day-cell
	// handler, which fires dateClick/select synchronously (event-calendar runs
	// it on pointerup, before bits-ui's own ~10ms-debounced outside-click
	// detection even gets a chance to run). Intercept it ourselves, on capture,
	// directly on the calendar element, before it ever gets there.
	let calendarWrapperEl: HTMLElement | null = $state(null);
	function onCalendarPointerDownCapture(e: PointerEvent) {
		if (!previewOpen) return;
		if ((e.target as HTMLElement | null)?.closest('[data-popover-content]')) return;
		previewOpen = false;
		e.stopPropagation();
	}
	$effect(() => {
		const el = calendarWrapperEl;
		if (!el) return;
		el.addEventListener('pointerdown', onCalendarPointerDownCapture, true);
		return () => el.removeEventListener('pointerdown', onCalendarPointerDownCapture, true);
	});

	function showPreview(kind: 'task' | 'event', data: Task | CalEvent, jsEvent: MouseEvent) {
		const { clientX, clientY } = jsEvent;
		previewAnchor = { getBoundingClientRect: () => new DOMRect(clientX, clientY, 0, 0) };
		previewItem = { kind, data } as typeof previewItem;
		previewOpen = true;
	}

	let contentEl = $state<HTMLElement | null>(null);

	function toDateISO(d: Date): string {
		return new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate())).toISOString();
	}

	// ── EC computed events ────────────────────────────────────────────────────
	const ecEvents = $derived((() => {
		const filteredEvents = events.filter(ev => {
			if (ev.birthday_of ? (filterTypes.size > 0 && !filterTypes.has('birthday')) : (filterTypes.size > 0 && !filterTypes.has('event'))) return false;
			if (filterMemberIDs.length > 0 && !filterMemberIDs.some(id => ev.attendee_ids?.includes(id))) return false;
			if (filterCategoryIDs.size > 0 && (!ev.category_id || !filterCategoryIDs.has(ev.category_id))) return false;
			return true;
		});
		const filteredTasks = filterTypes.size > 0 && !filterTypes.has('task') ? [] : tasks.filter(t => {
			if (!t.end_date) return false;
			if (t.status === 'done' && !showCompleted) return false;
			if (filterMemberIDs.length > 0 && !filterMemberIDs.includes(t.assigned_to ?? '')) return false;
			if (filterCategoryIDs.size > 0 && (!t.category_id || !filterCategoryIDs.has(t.category_id))) return false;
			return true;
		});
		return [
			...filteredEvents.map(ev => {
				const hex = ev.birthday_of ? '#ec4899' : categoryHex(ev.category_id);
				const prefix = ev.birthday_of ? '🎂 ' : '';
				const star = ev.important ? '★ ' : '';
				const title = star + prefix + ev.title;
				return {
					id: ev.id, title,
					start: ev.all_day ? ev.start_at : new Date(ev.start_at),
					end: ev.all_day ? ev.end_at : new Date(ev.end_at),
					allDay: ev.all_day,
					startEditable: !ev.birthday_of,
					durationEditable: !ev.birthday_of,
					...(hex ? { backgroundColor: hex, borderColor: hex, textColor: '#fff' } : {}),
					extendedProps: { type: 'event', data: ev },
				};
			}),
			...filteredTasks.map(t => {
				const done = t.status === 'done';
				const hex = done ? null : categoryHex(t.category_id);
				const hasTime = taskHasTime(t.end_date!);
				const start = hasTime ? new Date(t.end_date!) : t.end_date;
				const end = hasTime ? new Date(new Date(t.end_date!).getTime() + 60 * 60 * 1000) : t.end_date;
				return {
					id: `task-${t.id}`, title: (t.important ? '★ ' : '') + t.title, start, end, allDay: !hasTime,
					startEditable: !done, durationEditable: false,
					...(hex ? { backgroundColor: hex, borderColor: hex, textColor: '#fff' } : {}),
					classNames: done ? ['ec-task', 'ec-task-done'] : ['ec-task'],
					extendedProps: { type: 'task', data: t },
				};
			}),
		];
	})());

	$effect(() => { ecOptions.events = ecEvents; });

	$effect(() => {
		const format = { hour: 'numeric', minute: '2-digit', ...hour12Option() };
		ecOptions.eventTimeFormat = format;
		ecOptions.slotLabelFormat = format;
	});

	// ── EC options ────────────────────────────────────────────────────────────
	function timeGridEventContent({ event }: any) {
		const h4 = document.createElement('h4');
		h4.className = 'ec-event-title';
		h4.textContent = event.title;
		return { domNodes: [h4] };
	}

	let ecOptions = $state<Record<string, unknown>>({
		view: 'dayGridMonth',
		date: today,
		height: '100%',
		headerToolbar: false,
		nowIndicator: true,
		selectable: true,
		editable: true,
		scrollTime: '08:00:00',
		firstDay: 1,
		dayMaxEvents: true,
		events: [],
		views: {
			timeGridWeek: { eventContent: timeGridEventContent },
			timeGridDay: { eventContent: timeGridEventContent },
		},
		datesSet: ({ view }: any) => {
			periodLabel = view.title;
			viewStart    = view.activeStart;
			viewEnd      = view.activeEnd;
			currentStart = view.currentStart;
			loadData(view.activeStart, view.activeEnd);
		},
		eventClick: ({ event, jsEvent }: any) => {
			if (event.extendedProps.type === 'event') {
				const data = event.extendedProps.data as CalEvent;
				if (data.subscription_id) {
					toast.info(msg.calendar_synced_readonly());
					return;
				}
				showPreview('event', data, jsEvent);
			}
			else if (event.extendedProps.type === 'task') showPreview('task', event.extendedProps.data as Task, jsEvent);
		},
		eventDrop: async ({ event, revert }: any) => {
			try {
				if (event.extendedProps.type === 'event' && (event.extendedProps.data as CalEvent).subscription_id) {
					revert();
					return;
				}
				if (event.extendedProps.type === 'task') {
					const t = event.extendedProps.data as Task;
					const newDate = event.allDay ? toDateISO(event.start as Date) : (event.start as Date).toISOString();
					await api.patch(`/api/v1/households/${familyID}/tasks/${t.id}`, {
						title: t.title, description: t.description, important: t.important,
						status: t.status, assigned_to: t.assigned_to, category_id: t.category_id,
						end_date: newDate,
					});
					tasks = tasks.map(tk => tk.id === t.id ? { ...tk, end_date: newDate } : tk);
				} else {
					await patchEvent(event.extendedProps.data as CalEvent, event.start, event.end ?? event.start, event.allDay);
					await loadData(viewStart, viewEnd);
				}
			} catch { revert(); }
		},
		eventResize: async ({ event, revert }: any) => {
			if (event.extendedProps.type === 'event' && (event.extendedProps.data as CalEvent).subscription_id) {
				revert();
				return;
			}
			try {
				await patchEvent(event.extendedProps.data as CalEvent, event.start, event.end, event.allDay);
				await loadData(viewStart, viewEnd);
			} catch { revert(); }
		},
		dateClick: ({ date, allDay }: any) => {
			createDialog?.open('event', date, date, allDay);
		},
		select: ({ start, end, allDay }: any) => {
			const s = start as Date;
			// allDay select: end is exclusive (next day), clamp back one ms
			const e = allDay ? new Date((end as Date).getTime() - 1) : end as Date;
			createDialog?.open('event', s, e, allDay);
		},
	});

	// ── Data loading ──────────────────────────────────────────────────────────
	async function loadData(from: Date, to: Date) {
		try {
			const [evs, tsks, mems, cats] = await Promise.all([
				api.get<CalEvent[]>(`/api/v1/households/${familyID}/events?from=${from.toISOString()}&to=${to.toISOString()}`).then(r => r ?? []),
				api.get<Task[]>(`/api/v1/households/${familyID}/tasks`).then(r => r ?? []),
				members.length ? Promise.resolve(members) : api.get<Member[]>(`/api/v1/households/${familyID}/members`).then(r => r ?? []),
				api.get<AppCategory[]>(`/api/v1/households/${familyID}/categories`).then(r => r ?? []),
			]);
			events = evs; tasks = tsks; members = mems; categories = cats;
		} catch { }
	}

	// ── SSE ───────────────────────────────────────────────────────────────────
	let es: EventSource | null = null;
	onMount(async () => {
		if (window.innerWidth < 768) appView = 'agenda';
		if (appView === 'agenda') {
			await loadAgenda();
		}
		es = new EventSource(sseUrl(`/api/v1/households/${familyID}/stream`));
		es.onmessage = (e) => {
			if (e.data !== 'refresh') return;
			if (appView === 'agenda') loadAgenda();
			else loadData(viewStart, viewEnd);
		};
		es.onerror = () => { es?.close(); es = null; };
	});
	onDestroy(() => es?.close());

	// ── Navigation ────────────────────────────────────────────────────────────
	function switchView(v: AppView) {
		appView = v;
		if (v === 'agenda') {
			periodLabel = msg.cal_upcoming();
			agendaReady = false;
			loadAgenda();
		} else {
			ecOptions.view = EC_VIEW[v];
		}
	}

	function prevPeriod() {
		const d = new Date(currentStart);
		if (appView === 'month') d.setMonth(d.getMonth() - 1);
		else if (appView === 'week') d.setDate(d.getDate() - 7);
		else d.setDate(d.getDate() - 1);
		ecOptions.date = d;
	}
	function nextPeriod() {
		const d = new Date(currentStart);
		if (appView === 'month') d.setMonth(d.getMonth() + 1);
		else if (appView === 'week') d.setDate(d.getDate() + 7);
		else d.setDate(d.getDate() + 1);
		ecOptions.date = d;
	}
	function jumpToToday() { ecOptions.date = new Date(today); }

	let jumpOpen = $state(false);
	const jumpValue = $derived<DateValue>(
		new CalendarDate(currentStart.getFullYear(), currentStart.getMonth() + 1, currentStart.getDate())
	);
	function jumpTo(v: DateValue | undefined) {
		if (!v) return;
		ecOptions.date = new Date(v.year, v.month - 1, v.day);
		jumpOpen = false;
	}

	// ── Event / Task CRUD ──────────────────────────────────────────────────────
	async function patchEvent(ev: CalEvent, start: Date, end: Date, allDay: boolean) {
		await api.patch(`/api/v1/households/${familyID}/events/${ev.id}`, {
			title: ev.title, description: ev.description ?? '', location: ev.location ?? '',
			start_at: allDay ? toDateISO(start) : start.toISOString(),
			end_at: allDay ? toDateISO(end) : end.toISOString(),
			all_day: allDay, attendee_ids: ev.attendee_ids ?? [], category_id: ev.category_id,
		});
	}

	async function toggleTask(task: Task, e: MouseEvent) {
		e.stopPropagation();
		const newStatus = task.status === 'done' ? 'todo' : 'done';
		try {
			await api.patch(`/api/v1/households/${familyID}/tasks/${task.id}`, {
				title: task.title, description: task.description, important: task.important,
				status: newStatus, assigned_to: task.assigned_to, end_date: task.end_date, category_id: task.category_id,
			});
			tasks = tasks.map(t => t.id === task.id ? { ...t, status: newStatus } : t);
		} catch { }
	}
</script>

<div class="h-full flex flex-col">

<!-- Header -->
<div class="shrink-0 px-4 md:px-6 pt-4 md:pt-6 pb-2">
<div class="flex flex-col sm:flex-row sm:items-center gap-2 mb-3">
	<div class="flex rounded-md border border-border overflow-hidden text-sm shrink-0 self-start">
		{#each [['month','M',msg.cal_view_month()],['week','W',msg.cal_view_week()],['day','D',msg.cal_view_day()],['agenda','A',msg.cal_view_agenda()]] as [v, short, long]}
			<button
				onclick={() => switchView(v as AppView)}
				class="px-2.5 py-1.5 transition-colors cursor-pointer {appView === v ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted'}"
			>
				<span class="sm:hidden">{short}</span>
				<span class="hidden sm:inline">{long}</span>
			</button>
		{/each}
	</div>

	<div class="hidden sm:block w-px h-5 bg-border shrink-0"></div>

	{#if appView !== 'agenda'}
		<div class="flex items-center gap-2">
			<Button variant="outline" size="sm" onclick={jumpToToday}>{msg.nav_today()}</Button>
			<div class="flex items-center rounded-md border border-border overflow-hidden shrink-0">
				<button onclick={prevPeriod} aria-label={msg.cal_previous_aria()} class="h-8 w-8 flex items-center justify-center hover:bg-muted transition-colors cursor-pointer">‹</button>
				<div class="w-px self-stretch bg-border"></div>
				<button onclick={nextPeriod} aria-label={msg.cal_next_aria()} class="h-8 w-8 flex items-center justify-center hover:bg-muted transition-colors cursor-pointer">›</button>
			</div>
			<Popover.Root bind:open={jumpOpen}>
				<Popover.Trigger class="h-8 pl-2.5 pr-2 flex items-center gap-1 rounded-md border border-border text-sm font-medium hover:bg-muted transition-colors cursor-pointer">
					<span class="w-32 sm:w-40 truncate text-left">{periodLabel}</span>
					<ChevronDown class="w-3.5 h-3.5 text-muted-foreground shrink-0" />
				</Popover.Trigger>
				<Popover.Content class="w-auto p-0" align="start">
					<DatePicker type="single" value={jumpValue} onValueChange={jumpTo} captionLayout="dropdown" />
				</Popover.Content>
			</Popover.Root>
		</div>
	{:else}
		<Button variant="outline" size="sm" onclick={scrollToToday} class="self-start">{msg.nav_today()}</Button>
	{/if}
</div>

<!-- Legend / filter bar -->
<div class="flex items-center gap-2.5 mb-2 flex-wrap">
	<div class="flex items-center gap-1.5 flex-wrap">
		<button onclick={() => toggleType('task')} class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs transition-all cursor-pointer {chipCls(filterTypes.has('task'))}">
			<span class="w-2.5 h-2.5 rounded-full border border-dashed border-current shrink-0"></span>
			{msg.board_filter_tasks()}
		</button>
		<button onclick={() => toggleType('event')} class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs transition-all cursor-pointer {chipCls(filterTypes.has('event'))}">
			<span class="w-2.5 h-2.5 rounded-full bg-current shrink-0"></span>
			{msg.board_filter_events()}
		</button>
		<button onclick={() => toggleType('birthday')} class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs transition-all cursor-pointer {chipCls(filterTypes.has('birthday'))}">
			🎂 {msg.board_filter_birthdays()}
		</button>
	</div>

	<div class="w-px h-4 bg-border shrink-0"></div>

	<button onclick={() => (showCompleted = !showCompleted)} class="px-2.5 py-1 rounded-full text-xs transition-all cursor-pointer {showCompleted ? 'ring-1 ring-foreground opacity-100' : 'opacity-70 hover:opacity-100'}">
		{showCompleted ? msg.cal_hide_completed() : msg.cal_show_completed()}
	</button>

{#if categories.length > 0}
		<div class="w-px h-4 bg-border hidden sm:block shrink-0"></div>
		<div class="hidden sm:flex items-center gap-1.5 flex-wrap">
			{#each categories as cat (cat.id)}
				<button onclick={() => toggleCategory(cat.id)} class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs transition-all cursor-pointer {chipCls(filterCategoryIDs.has(cat.id))}">
					<span class="w-2 h-2 rounded-full {dotClass(cat.color)} shrink-0"></span>
					{cat.name}
				</button>
			{/each}
		</div>
	{/if}
	{#if members.length > 0}
		<div class="w-px h-4 bg-border hidden sm:block shrink-0"></div>
		<div class="hidden sm:flex items-center gap-1.5 flex-wrap">
			{#each members as m (m.user_id)}
				<button onclick={() => toggleMember(m.user_id)} title={m.name} class="rounded-full transition-all cursor-pointer shrink-0
					{filterMemberIDs.includes(m.user_id) ? 'ring-2 ring-primary ring-offset-1 opacity-100' : someFilterActive ? 'opacity-30' : 'opacity-80 hover:opacity-100'}">
					<UserAvatar name={m.name} avatarUrl={m.avatar_url} userId={m.user_id} size={24} />
				</button>
			{/each}
		</div>
	{/if}
	{#if someFilterActive}
		<button onclick={clearFilters} class="flex items-center gap-1 px-2.5 py-1 rounded-full text-xs text-muted-foreground hover:text-foreground hover:bg-muted transition-colors cursor-pointer">
			<X class="w-3 h-3" />{msg.board_clear_filters()}
		</button>
	{/if}
	{#if appView !== 'agenda'}
		<span class="text-xs text-muted-foreground/40 ml-auto hidden sm:block select-none">{msg.cal_click_drag_hint()}</span>
	{/if}
</div>

</div>

<div bind:this={contentEl} class="flex-1 min-h-0 overflow-auto">
{#if appView === 'agenda'}
	<div class="px-4 md:px-6 pb-8">
		{#if !agendaReady}
			<div class="flex items-center justify-center py-16">
				<span class="text-xs text-muted-foreground/50">{msg.invite_loading()}</span>
			</div>
		{:else}
		<!-- Load past button — explicit, no scroll detection -->
		<div class="flex justify-center mb-4">
			<Button variant="ghost" size="sm" onclick={extendAgendaBack} disabled={agendaLoadingTop}
				class="text-xs text-muted-foreground">
				{agendaLoadingTop ? msg.invite_loading() : `↑ ${msg.cal_load_previous_month()}`}
			</Button>
		</div>

		{#if agendaGroups.length === 0 && !agendaLoadingBottom}
			<div class="flex flex-col items-center gap-2 py-16 text-muted-foreground">
				<CalendarDays class="w-10 h-10 opacity-30" />
				<p class="text-sm font-medium">{msg.cal_nothing_here()}</p>
			</div>
		{:else}
			{#each agendaGroups as group (group.dayMs)}
				<!-- Day header -->
				<div data-agenda-today={group.dayMs === todayMs ? 'true' : undefined}
					class="flex items-center gap-3 mt-5 first:mt-0 mb-2">
					<span class="text-xs font-semibold uppercase tracking-wide shrink-0
						{group.dayMs === todayMs ? 'text-primary' : 'text-muted-foreground'}">
						{group.label}
					</span>
					<div class="flex-1 h-px bg-border"></div>
				</div>

				<!-- Events -->
				{#if group.events.length > 0}
					<div class="flex flex-col gap-0.5">
						{#each group.events as ev}
							{@const cat = categories.find(c => c.id === ev.category_id)}
							<button
								onclick={() => editDialog?.openEvent(ev)}
								class="flex items-baseline gap-3 text-left py-1 px-2 -mx-2 rounded-md hover:bg-accent/50 transition-colors cursor-pointer w-full"
							>
								<span class="text-xs text-muted-foreground tabular-nums w-12 shrink-0 text-right">
									{ev.all_day ? msg.today_all_day() : fmtTime(ev.start_at)}
								</span>
								{#if ev.birthday_of}<span class="text-sm shrink-0">🎂</span>{/if}
								<span class="text-sm font-medium flex-1 min-w-0 truncate">{ev.title}</span>
								{#if ev.location}
									<span class="text-xs text-muted-foreground truncate hidden sm:block max-w-32">{ev.location}</span>
								{/if}
								{#if cat}
									<span class="flex items-center gap-1 shrink-0">
										<span class="w-1.5 h-1.5 rounded-full {dotClass(cat.color)}"></span>
										<span class="text-xs text-muted-foreground hidden sm:block">{cat.name}</span>
									</span>
								{/if}
							</button>
						{/each}
					</div>
				{/if}

				<!-- Tasks (with divider if events also present) -->
				{#if group.tasks.length > 0}
					{#if group.events.length > 0}
						<div class="flex items-center gap-2 my-1.5 ml-[3.75rem]">
							<div class="flex-1 h-px bg-border/50"></div>
							<span class="text-[10px] uppercase tracking-wider text-muted-foreground/50 shrink-0">{msg.board_filter_tasks()}</span>
						</div>
					{/if}
					<div class="flex flex-col gap-0.5">
						{#each group.tasks as task}
							{@const cat = categories.find(c => c.id === task.category_id)}
							<button
								onclick={() => editDialog?.openTask(task)}
								class="flex items-center gap-3 text-left py-1 px-2 -mx-2 rounded-md hover:bg-accent/50 transition-colors cursor-pointer w-full"
							>
								<span class="w-12 shrink-0 flex justify-end">
									<span class="w-3.5 h-3.5 rounded-sm border-2 border-muted-foreground/30 shrink-0"></span>
								</span>
								<span class="text-sm flex-1 min-w-0 truncate {task.important ? 'font-medium' : ''} {task.status === 'done' ? 'line-through text-muted-foreground' : ''}">{task.title}</span>
								{#if cat}
									<span class="flex items-center gap-1 shrink-0">
										<span class="w-1.5 h-1.5 rounded-full {dotClass(cat.color)}"></span>
										<span class="text-xs text-muted-foreground hidden sm:block">{cat.name}</span>
									</span>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			{/each}
		{/if}
		<!-- Bottom sentinel for future scroll -->
		<div bind:this={agendaBottomSentinel} class="h-10 flex items-center justify-center mt-4">
			{#if agendaLoadingBottom}
				<span class="text-xs text-muted-foreground/50">{msg.invite_loading()}</span>
			{/if}
		</div>
		{/if}
	</div>
{:else}
	<!-- EC calendar for month / week / day -->
	<div class="px-4 md:px-6 pb-4 md:pb-6 h-full">
		<div bind:this={calendarWrapperEl} class="rounded-lg overflow-hidden h-full border border-border">
			<Calendar plugins={[DayGrid, TimeGrid, Interaction]} options={ecOptions} />
		</div>
	</div>
{/if}
</div>

</div>

<EditDialog
	bind:this={editDialog}
	{familyID} {members} {categories}
	onSaved={() => appView === 'agenda' ? loadAgenda() : loadData(viewStart, viewEnd)}
	onDeleted={() => appView === 'agenda' ? loadAgenda() : loadData(viewStart, viewEnd)}
/>
<CreateDialog
	bind:this={createDialog}
	{familyID} {members} {categories}
	onCreated={() => appView === 'agenda' ? loadAgenda() : loadData(viewStart, viewEnd)}
/>

<Popover.Root bind:open={previewOpen}>
	<Popover.Content customAnchor={previewAnchor} class="w-80 relative">
		{#if previewItem}
			{@const openEdit = () => {
				previewOpen = false;
				previewItem!.kind === 'event' ? editDialog?.openEvent(previewItem!.data as CalEvent) : editDialog?.openTask(previewItem!.data as Task);
			}}
			<div class="absolute top-2 right-2 flex items-center gap-0.5">
				<Button variant="ghost" size="icon" class="size-7 text-muted-foreground hover:text-foreground" onclick={openEdit} aria-label={msg.action_edit()}>
					<Pencil class="size-3.5" />
				</Button>
				<Button
					variant="ghost"
					size="icon"
					class="size-7 text-muted-foreground hover:text-destructive"
					aria-label={msg.edit_dialog_delete()}
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


