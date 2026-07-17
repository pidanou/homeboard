<script lang="ts">
	import '@event-calendar/core/index.css';
	import { Calendar, TimeGrid, Interaction } from '@event-calendar/core';
	import type { CalEvent, Task, AppCategory } from '$lib/types';
	import { CATEGORY_HEX } from '$lib/categories';
	import { taskHasTime, hour12Option } from '$lib/dates';
	import { getLocale } from '$lib/paraglide/runtime';

	type Props = {
		events: CalEvent[];
		tasks: Task[];
		categories: AppCategory[];
		onEventClick?: (event: CalEvent, jsEvent: MouseEvent) => void;
		onTaskClick?: (task: Task, jsEvent: MouseEvent) => void;
		onDateClick?: (date: Date, allDay: boolean) => void;
		onSelect?: (start: Date, end: Date, allDay: boolean) => void;
		onEventDrop?: (event: CalEvent, start: Date, end: Date, allDay: boolean, revert: () => void) => void;
		onTaskDrop?: (task: Task, newDate: Date, revert: () => void) => void;
	};

	const { events, tasks, categories, onEventClick, onTaskClick, onDateClick, onSelect, onEventDrop, onTaskDrop }: Props = $props();

	function categoryHex(categoryID: string | undefined): string | null {
		if (!categoryID) return null;
		const cat = categories.find(c => c.id === categoryID);
		return cat ? (CATEGORY_HEX[cat.color] ?? null) : null;
	}

	const ecEvents = $derived([
		...events.map(ev => {
			const hex = ev.birthday_of ? '#ec4899' : categoryHex(ev.category_id);
			const prefix = ev.birthday_of ? '🎂 ' : '';
			const star = ev.important ? '★ ' : '';
			return {
				id: ev.id, title: star + prefix + ev.title,
				start: ev.all_day ? ev.start_at : new Date(ev.start_at),
				end: ev.all_day ? ev.end_at : new Date(ev.end_at),
				allDay: ev.all_day,
				...(hex ? { backgroundColor: hex, borderColor: hex, textColor: '#fff' } : {}),
				extendedProps: { type: 'event', data: ev },
			};
		}),
		...tasks.filter(t => !!t.end_date && taskHasTime(t.end_date!)).map(t => {
			const done = t.status === 'done';
			const hex = done ? null : categoryHex(t.category_id);
			const start = new Date(t.end_date!);
			const end = new Date(start.getTime() + 60 * 60 * 1000);
			return {
				id: `task-${t.id}`, title: (t.important ? '★ ' : '') + t.title,
				start, end, allDay: false,
				...(hex ? { backgroundColor: hex, borderColor: hex, textColor: '#fff' } : {}),
				classNames: done ? ['ec-task', 'ec-task-done'] : ['ec-task'],
				extendedProps: { type: 'task', data: t },
			};
		}),
	]);

	const today = new Date();

	const ecOptions = $state({
		view: 'timeGridDay',
		date: today,
		height: '100%',
		locale: getLocale(),
		headerToolbar: { start: '', center: '', end: '' },
		nowIndicator: true,
		scrollTime: '08:00:00',
		firstDay: 1,
		selectable: true,
		editable: true,
		events: [] as unknown[],
		eventTimeFormat: { hour: 'numeric', minute: '2-digit' } as Record<string, unknown>,
		slotLabelFormat: { hour: 'numeric', minute: '2-digit' } as Record<string, unknown>,
		dateClick: ({ date, allDay }: any) => onDateClick?.(date, allDay),
		select: ({ start, end, allDay }: any) => onSelect?.(start, end, allDay),
		eventDrop: ({ event, revert }: any) => {
			if (event.extendedProps.type === 'task') onTaskDrop?.(event.extendedProps.data, event.start, revert);
			else onEventDrop?.(event.extendedProps.data, event.start, event.end ?? event.start, event.allDay, revert);
		},
		eventResize: ({ event, revert }: any) => {
			onEventDrop?.(event.extendedProps.data, event.start, event.end, event.allDay, revert);
		},
		eventClick: ({ event, jsEvent }: any) => {
			if (event.extendedProps.type === 'event') onEventClick?.(event.extendedProps.data as CalEvent, jsEvent);
			else if (event.extendedProps.type === 'task') onTaskClick?.(event.extendedProps.data as Task, jsEvent);
		},
	});

	$effect(() => { ecOptions.events = ecEvents; });

	$effect(() => {
		const format = { hour: 'numeric', minute: '2-digit', ...hour12Option() };
		ecOptions.eventTimeFormat = format;
		ecOptions.slotLabelFormat = format;
	});

	let isDark = $state(false);
	$effect(() => {
		const update = () => { isDark = document.documentElement.classList.contains('dark'); };
		update();
		const obs = new MutationObserver(update);
		obs.observe(document.documentElement, { attributeFilter: ['class'] });
		return () => obs.disconnect();
	});
</script>

<div class="h-full min-h-[300px] ring-1 ring-border rounded-xl overflow-hidden" class:ec-dark={isDark}>
	<Calendar plugins={[TimeGrid, Interaction]} options={ecOptions} />
</div>
