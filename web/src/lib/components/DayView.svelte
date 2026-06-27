<script lang="ts">
	import '@event-calendar/core/index.css';
	import { Calendar, TimeGrid, Interaction } from '@event-calendar/core';
	import type { CalEvent, Task, AppCategory } from '$lib/types';
	import { CATEGORY_HEX } from '$lib/categories';
	import { taskHasTime } from '$lib/dates';

	type Props = {
		events: CalEvent[];
		tasks: Task[];
		categories: AppCategory[];
		onEventClick?: (event: CalEvent) => void;
		onTaskClick?: (task: Task) => void;
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
			const prefix = ev.birthday_of ? '🎂' : ev.icon;
			const star = ev.important ? '★ ' : '';
			return {
				id: ev.id, title: star + (prefix ? prefix + ' ' + ev.title : ev.title),
				start: ev.start_at, end: ev.end_at, allDay: ev.all_day,
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
				id: `task-${t.id}`, title: (t.important ? '★ ' : '') + (t.icon ? `${t.icon} ` : '') + t.title,
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
		headerToolbar: { start: '', center: '', end: '' },
		nowIndicator: true,
		scrollTime: '08:00:00',
		firstDay: 1,
		selectable: true,
		editable: true,
		events: [] as unknown[],
		dateClick: ({ date, allDay }: any) => onDateClick?.(date, allDay),
		select: ({ start, end, allDay }: any) => onSelect?.(start, end, allDay),
		eventDrop: ({ event, revert }: any) => {
			if (event.extendedProps.type === 'task') onTaskDrop?.(event.extendedProps.data, event.start, revert);
			else onEventDrop?.(event.extendedProps.data, event.start, event.end ?? event.start, event.allDay, revert);
		},
		eventResize: ({ event, revert }: any) => {
			onEventDrop?.(event.extendedProps.data, event.start, event.end, event.allDay, revert);
		},
		eventClick: ({ event }: any) => {
			if (event.extendedProps.type === 'event') onEventClick?.(event.extendedProps.data as CalEvent);
			else if (event.extendedProps.type === 'task') onTaskClick?.(event.extendedProps.data as Task);
		},
	});

	$effect(() => { ecOptions.events = ecEvents; });
</script>

<div class="ec-auto-dark h-full min-h-[300px] border border-border rounded-xl overflow-hidden">
	<Calendar plugins={[TimeGrid, Interaction]} options={ecOptions} />
</div>
