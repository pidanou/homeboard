<script lang="ts">
	import '@event-calendar/core/index.css';
	import { page } from '$app/stores';
	import { onMount, onDestroy, tick } from 'svelte';
	import { Calendar, DayGrid, TimeGrid, Interaction } from '@event-calendar/core';
	import { api, sseUrl } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { X, CalendarDays } from 'lucide-svelte';
	import type { CalEvent, Task, Member, AppCategory } from '$lib/types';
	import { dotClass, CATEGORY_HEX } from '$lib/categories';
	import { fmtTime } from '$lib/dates';
	import EditDialog from '$lib/components/EditDialog.svelte';
	import CreateDialog from '$lib/components/CreateDialog.svelte';

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
	let error = $state('');

	// Current visible date range (set by EC's datesSet callback)
	let viewStart    = $state(new Date(today.getFullYear(), today.getMonth(), 1));
	let viewEnd      = $state(new Date(today.getFullYear(), today.getMonth() + 1, 1));
	let currentStart = $state(new Date(today.getFullYear(), today.getMonth(), 1));

	// ── Filters ───────────────────────────────────────────────────────────────
	let filterTypes = $state(new Set<'task' | 'event'>());
	let filterMemberIDs = $state<string[]>([]);
	let filterCategoryID = $state<string | null>(null);
	const someFilterActive = $derived(filterTypes.size > 0 || filterMemberIDs.length > 0 || filterCategoryID !== null);

	function toggleType(t: 'task' | 'event') {
		const next = new Set(filterTypes);
		next.has(t) ? next.delete(t) : next.add(t);
		filterTypes = next;
	}
	function toggleMember(id: string) {
		filterMemberIDs = filterMemberIDs.includes(id) ? filterMemberIDs.filter(x => x !== id) : [...filterMemberIDs, id];
	}
	function toggleCategory(id: string) {
		filterCategoryID = filterCategoryID === id ? null : id;
	}
	function clearFilters() { filterTypes = new Set(); filterMemberIDs = []; filterCategoryID = null; }
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
				api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${agendaStart.toISOString()}&to=${agendaEnd.toISOString()}`).then(r => r ?? []),
				api.get<Task[]>(`/api/v1/families/${familyID}/tasks`).then(r => r ?? []),
				members.length ? Promise.resolve(members) : api.get<Member[]>(`/api/v1/families/${familyID}/members`).then(r => r ?? []),
				categories.length ? Promise.resolve(categories) : api.get<AppCategory[]>(`/api/v1/families/${familyID}/categories`).then(r => r ?? []),
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
			const evs = await api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${agendaEnd.toISOString()}&to=${newEnd.toISOString()}`).then(r => r ?? []);
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
			const scrollEl = document.querySelector('main');
			const prevHeight = scrollEl?.scrollHeight ?? 0;
			const prevTop    = scrollEl?.scrollTop ?? 0;
			const evs = await api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${newStart.toISOString()}&to=${agendaStart.toISOString()}`).then(r => r ?? []);
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
		const scrollEl = document.querySelector('main');
		if (!el || !scrollEl) return;
		scrollEl.scrollTo({ top: scrollEl.scrollTop + el.getBoundingClientRect().top - headerHeight - 16, behavior: 'smooth' });
	}

	type AgendaGroup = { dayMs: number; label: string; events: CalEvent[]; tasks: Task[] };

	const agendaGroups = $derived((() => {
		if (appView !== 'agenda') return [] as AgendaGroup[];
		const showEvents = filterTypes.size === 0 || filterTypes.has('event');
		const showTasks  = filterTypes.size === 0 || filterTypes.has('task');
		const byMemberEv = (ev: CalEvent) =>
			filterMemberIDs.length === 0 || (ev.attendee_ids ?? []).some(id => filterMemberIDs.includes(id));
		const byMember = (id: string | undefined) =>
			filterMemberIDs.length === 0 || (!!id && filterMemberIDs.includes(id));
		const byCat = (id: string | undefined) =>
			filterCategoryID === null || id === filterCategoryID;
		const startMs = agendaStart.getTime();
		const endMs   = agendaEnd.getTime();

		// Pre-fill every day in the loaded range — past days appear when explicitly loaded via button
		const dayMap = new Map<number, { evs: CalEvent[]; tsks: Task[] }>();
		for (let ms = startMs; ms <= endMs; ms += 86400000) {
			dayMap.set(ms, { evs: [], tsks: [] });
		}

		if (showEvents) {
			for (const ev of agendaEvents) {
				if (!byMemberEv(ev) || !byCat(ev.category_id)) continue;
				const d = new Date(ev.start_at);
				const dayMs = new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
				dayMap.get(dayMs)?.evs.push(ev);
			}
		}
		if (showTasks) {
			for (const t of tasks) {
				if (t.status === 'done' || !t.end_date) continue;
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
				label: new Date(dayMs).toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' }),
				events: evs.sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime()),
				tasks: tsks,
			}));
	})());

	// ── Dialogs ───────────────────────────────────────────────────────────────
	let editDialog: { openTask: (t: Task) => void; openEvent: (e: CalEvent) => void } | undefined = $state();
	let createDialog: { open: (t?: 'task' | 'event') => void } | undefined = $state();

	// ── Header height (for dynamic calendar height) ───────────────────────────
	let headerEl = $state<HTMLElement | null>(null);
	let headerHeight = $state(160);
	$effect(() => {
		if (!headerEl) return;
		const ro = new ResizeObserver(() => { headerHeight = headerEl!.offsetHeight; });
		ro.observe(headerEl);
		return () => ro.disconnect();
	});

	function toDateISO(d: Date): string {
		return new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate())).toISOString();
	}

	// ── EC computed events ────────────────────────────────────────────────────
	const ecEvents = $derived((() => {
		const filteredEvents = filterTypes.size > 0 && !filterTypes.has('event') ? [] : events.filter(ev => {
			if (filterMemberIDs.length > 0 && !filterMemberIDs.some(id => ev.attendee_ids?.includes(id))) return false;
			if (filterCategoryID !== null && ev.category_id !== filterCategoryID) return false;
			return true;
		});
		const filteredTasks = filterTypes.size > 0 && !filterTypes.has('task') ? [] : tasks.filter(t => {
			if (t.status === 'done' || !t.end_date) return false;
			if (filterMemberIDs.length > 0 && !filterMemberIDs.includes(t.assigned_to ?? '')) return false;
			if (filterCategoryID !== null && t.category_id !== filterCategoryID) return false;
			return true;
		});
		return [
			...filteredEvents.map(ev => {
				const hex = categoryHex(ev.category_id);
				return {
					id: ev.id, title: ev.title, start: ev.start_at, end: ev.end_at, allDay: ev.all_day,
					editable: true,
					...(hex ? { backgroundColor: hex, borderColor: hex, textColor: '#fff' } : {}),
					extendedProps: { type: 'event', data: ev },
				};
			}),
			...filteredTasks.map(t => {
				const hex = categoryHex(t.category_id) ?? '#94a3b8';
				return {
					id: `task-${t.id}`, title: t.title, start: t.end_date, end: t.end_date, allDay: true,
					startEditable: true, durationEditable: false,
					backgroundColor: hex + '18', borderColor: hex, textColor: 'var(--foreground)',
					classNames: ['ec-task'],
					extendedProps: { type: 'task', data: t },
				};
			}),
		];
	})());

	$effect(() => { ecOptions.events = ecEvents; });

	// ── EC options ────────────────────────────────────────────────────────────
	let ecOptions = $state<Record<string, unknown>>({
		view: 'dayGridMonth',
		date: today,
		height: '100%',
		headerToolbar: { start: '', center: '', end: '' },
		nowIndicator: true,
		selectable: true,
		editable: true,
		scrollTime: '08:00:00',
		firstDay: 1,
		dayMaxEvents: true,
		events: [],
		datesSet: ({ view }: any) => {
			periodLabel = view.title;
			viewStart    = view.activeStart;
			viewEnd      = view.activeEnd;
			currentStart = view.currentStart;
			loadData(view.activeStart, view.activeEnd);
		},
		eventClick: ({ event }: any) => {
			if (event.extendedProps.type === 'event') editDialog?.openEvent(event.extendedProps.data as CalEvent);
			else if (event.extendedProps.type === 'task') editDialog?.openTask(event.extendedProps.data as Task);
		},
		eventDrop: async ({ event, revert }: any) => {
			try {
				if (event.extendedProps.type === 'task') {
					const t = event.extendedProps.data as Task;
					const newDate = toDateISO(event.start as Date);
					await api.patch(`/api/v1/families/${familyID}/tasks/${t.id}`, {
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
			try {
				await patchEvent(event.extendedProps.data as CalEvent, event.start, event.end, event.allDay);
				await loadData(viewStart, viewEnd);
			} catch { revert(); }
		},
		dateClick: () => { createDialog?.open(); },
		select: () => { createDialog?.open(); },
	});

	// ── Data loading ──────────────────────────────────────────────────────────
	async function loadData(from: Date, to: Date) {
		try {
			const [evs, tsks, mems, cats] = await Promise.all([
				api.get<CalEvent[]>(`/api/v1/families/${familyID}/events?from=${from.toISOString()}&to=${to.toISOString()}`).then(r => r ?? []),
				api.get<Task[]>(`/api/v1/families/${familyID}/tasks`).then(r => r ?? []),
				members.length ? Promise.resolve(members) : api.get<Member[]>(`/api/v1/families/${familyID}/members`).then(r => r ?? []),
				categories.length ? Promise.resolve(categories) : api.get<AppCategory[]>(`/api/v1/families/${familyID}/categories`).then(r => r ?? []),
			]);
			events = evs; tasks = tsks; members = mems; categories = cats;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load calendar';
		}
	}

	// ── SSE ───────────────────────────────────────────────────────────────────
	let es: EventSource | null = null;
	onMount(async () => {
		if (window.innerWidth < 768) appView = 'agenda';
		if (appView === 'agenda') {
			await loadAgenda();
		}
		es = new EventSource(sseUrl(`/api/v1/families/${familyID}/stream`));
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
			periodLabel = 'Upcoming';
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

	// ── Event / Task CRUD ──────────────────────────────────────────────────────
	async function patchEvent(ev: CalEvent, start: Date, end: Date, allDay: boolean) {
		await api.patch(`/api/v1/families/${familyID}/events/${ev.id}`, {
			title: ev.title, description: ev.description ?? '', location: ev.location ?? '',
			start_at: start.toISOString(), end_at: end.toISOString(),
			all_day: allDay, attendee_ids: ev.attendee_ids ?? [], category_id: ev.category_id,
		});
	}

	async function toggleTask(task: Task, e: MouseEvent) {
		e.stopPropagation();
		const newStatus = task.status === 'done' ? 'todo' : 'done';
		try {
			await api.patch(`/api/v1/families/${familyID}/tasks/${task.id}`, {
				title: task.title, description: task.description, important: task.important,
				status: newStatus, assigned_to: task.assigned_to, end_date: task.end_date, category_id: task.category_id,
			});
			tasks = tasks.map(t => t.id === task.id ? { ...t, status: newStatus } : t);
		} catch (err) { error = err instanceof Error ? err.message : 'Something went wrong'; }
	}
</script>

{#if error}
	<p class="text-sm text-destructive mb-2 px-4 md:px-6">{error}</p>
{/if}

<!-- Header -->
<div bind:this={headerEl} class="sticky top-0 z-10 bg-background px-4 md:px-6 pt-4 md:pt-6 pb-2">
<div class="flex items-center justify-between mb-3 gap-2 flex-wrap">
	<div class="flex items-center gap-1.5 flex-wrap">
		<div class="flex rounded-md border border-border overflow-hidden text-sm shrink-0">
			{#each [['month','M','Month'],['week','W','Week'],['day','D','Day'],['agenda','A','Agenda']] as [v, short, long]}
				<button
					onclick={() => switchView(v as AppView)}
					class="px-2.5 py-1.5 transition-colors cursor-pointer {appView === v ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted'}"
				>
					<span class="sm:hidden">{short}</span>
					<span class="hidden sm:inline">{long}</span>
				</button>
			{/each}
		</div>

		{#if appView !== 'agenda'}
			<Button variant="outline" size="sm" onclick={prevPeriod} aria-label="Previous">‹</Button>
			<span class="text-sm font-medium max-w-40 truncate text-center">{periodLabel}</span>
			<Button variant="outline" size="sm" onclick={nextPeriod} aria-label="Next">›</Button>
			<Button variant="outline" size="sm" onclick={jumpToToday}>Today</Button>
		{:else}
			<Button variant="outline" size="sm" onclick={scrollToToday}>Today</Button>
		{/if}
	</div>
</div>

<!-- Legend / filter bar -->
<div class="flex items-center gap-2 mb-2 flex-wrap">
	<button onclick={() => toggleType('task')} class="flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs transition-all cursor-pointer {chipCls(filterTypes.has('task'))}">
		<span class="inline-flex items-center justify-center w-3 h-3 rounded-sm border-2 border-current shrink-0"></span>
		Tasks
	</button>
	<button onclick={() => toggleType('event')} class="flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs transition-all cursor-pointer {chipCls(filterTypes.has('event'))}">
		<span class="inline-block w-3 h-3 rounded-sm bg-current shrink-0"></span>
		Events
	</button>
	{#if categories.length > 0}
		<span class="text-border text-xs hidden sm:block">|</span>
		{#each categories as cat (cat.id)}
			<button onclick={() => toggleCategory(cat.id)} class="hidden sm:flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs transition-all cursor-pointer {chipCls(filterCategoryID === cat.id)}">
				<span class="w-2 h-2 rounded-full {dotClass(cat.color)} shrink-0"></span>
				{cat.name}
			</button>
		{/each}
	{/if}
	{#if members.length > 0}
		<span class="text-border text-xs hidden sm:block">|</span>
		{#each members as m (m.user_id)}
			<button onclick={() => toggleMember(m.user_id)} title={m.name} class="hidden sm:flex w-6 h-6 rounded-full text-[10px] font-semibold items-center justify-center transition-all cursor-pointer shrink-0
				{filterMemberIDs.includes(m.user_id) ? 'bg-primary text-primary-foreground ring-1 ring-foreground' : someFilterActive ? 'bg-muted text-muted-foreground opacity-30' : 'bg-muted text-muted-foreground opacity-70 hover:opacity-100'}">
				{initials(m.name)}
			</button>
		{/each}
	{/if}
	{#if someFilterActive}
		<button onclick={clearFilters} class="flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer ml-1">
			<X class="w-3 h-3" />Clear
		</button>
	{/if}
	{#if appView !== 'agenda'}
		<span class="text-xs text-muted-foreground/40 ml-auto hidden sm:block select-none">Click or drag to add</span>
	{/if}
</div>
</div>

{#if appView === 'agenda'}
	<div class="px-4 md:px-6 pb-8">
		{#if !agendaReady}
			<div class="flex items-center justify-center py-16">
				<span class="text-xs text-muted-foreground/50">Loading…</span>
			</div>
		{:else}
		<!-- Load past button — explicit, no scroll detection -->
		<div class="flex justify-center mb-4">
			<Button variant="ghost" size="sm" onclick={extendAgendaBack} disabled={agendaLoadingTop}
				class="text-xs text-muted-foreground">
				{agendaLoadingTop ? 'Loading…' : '↑ Load previous month'}
			</Button>
		</div>

		{#if agendaGroups.length === 0 && !agendaLoadingBottom}
			<div class="flex flex-col items-center gap-2 py-16 text-muted-foreground">
				<CalendarDays class="w-10 h-10 opacity-30" />
				<p class="text-sm font-medium">Nothing here</p>
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
									{ev.all_day ? 'All day' : fmtTime(ev.start_at)}
								</span>
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
							<span class="text-[10px] uppercase tracking-wider text-muted-foreground/50 shrink-0">Tasks</span>
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
								<span class="text-sm flex-1 min-w-0 truncate {task.important ? 'font-medium' : ''}">{task.title}</span>
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
				<span class="text-xs text-muted-foreground/50">Loading…</span>
			{/if}
		</div>
		{/if}
	</div>
{:else}
	<!-- EC calendar for month / week / day -->
	<div class="px-4 md:px-6" style="height: calc(100dvh - {headerHeight}px - 4rem)">
		<div class="rounded-lg overflow-hidden h-full border border-border">
			<Calendar plugins={[DayGrid, TimeGrid, Interaction]} options={ecOptions} />
		</div>
	</div>
{/if}

<EditDialog
	bind:this={editDialog}
	{familyID} {members} {categories}
	onSaved={() => appView === 'agenda' ? loadAgenda() : loadData(viewStart, viewEnd)}
	onDeleted={() => appView === 'agenda' ? loadAgenda() : loadData(viewStart, viewEnd)}
	onError={(e) => { error = e instanceof Error ? e.message : 'Something went wrong'; }}
/>
<CreateDialog
	bind:this={createDialog}
	{familyID} {members} {categories}
	onCreated={() => appView === 'agenda' ? loadAgenda() : loadData(viewStart, viewEnd)}
	onError={(e) => { error = e instanceof Error ? e.message : 'Something went wrong'; }}
/>


<style>
	/* Remove EC's empty toolbar — we use our own header */
	:global(.ec-toolbar) {
		display: none;
	}

	:global(.ec-event.ec-task) {
		border-width: 1px !important;
		border-left-width: 4px !important;
	}
	:global(.ec-day-grid .ec-body .ec-day),
	:global(.ec-time-grid .ec-body .ec-time) {
		cursor: pointer;
	}
	:global(.ec-time-grid .ec-today) {
		--ec-day-bg-color: var(--ec-bg-color);
		--ec-today-bg-color: var(--ec-bg-color);
	}

	/* Light mode EC theme mapped to our design system */
	:global(.ec) {
		--ec-bg-color: var(--background);
		--ec-text-color: var(--foreground);
		--ec-border-color: var(--border);
		--ec-event-bg-color: oklch(0.52 0.14 245);
		--ec-event-text-color: #fff;
		--ec-today-bg-color: color-mix(in oklch, var(--primary) 10%, var(--background));
		--ec-highlight-color: color-mix(in oklch, var(--primary) 6%, var(--background));
		--ec-now-indicator-color: oklch(0.63 0.24 25);
		--ec-popup-bg-color: var(--popover);
		font-family: inherit;
		font-size: 0.875rem;
		height: 100%;
	}

	/* Dark mode — EC doesn't know about our .dark class */
	:global(.dark .ec) {
		color-scheme: dark;
		--ec-color-400: oklch(43.9% 0 0);
		--ec-color-300: oklch(37.1% 0 0);
		--ec-color-200: oklch(26.9% 0 0);
		--ec-color-100: oklch(20.5% 0 0);
		--ec-color-50: oklch(14.5% 0 0);
		--ec-bg-color: var(--background);
		--ec-border-color: var(--border);
		--ec-today-bg-color: color-mix(in oklch, var(--primary) 15%, var(--background));
		--ec-highlight-color: color-mix(in oklch, var(--primary) 8%, var(--background));
		--ec-popup-bg-color: var(--popover);
	}
</style>
