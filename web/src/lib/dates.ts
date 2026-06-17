import { CalendarDate, type DateValue } from '@internationalized/date';
import type { DateRange } from 'bits-ui';

export function calDateToISO(d: DateValue): string {
	const pad = (n: number) => String(n).padStart(2, '0');
	return new Date(`${d.year}-${pad(d.month)}-${pad(d.day)}`).toISOString();
}

export function isoToCalDate(iso: string): CalendarDate {
	const d = new Date(iso);
	return new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate());
}

export function fmtCalDate(d: DateValue): string {
	return new Date(
		`${d.year}-${String(d.month).padStart(2, '0')}-${String(d.day).padStart(2, '0')}`,
	).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
}

export function calDateTimeToISO(d: DateValue, time: string, allDay: boolean): string {
	const pad = (n: number) => String(n).padStart(2, '0');
	return new Date(
		`${d.year}-${pad(d.month)}-${pad(d.day)}T${allDay ? '00:00' : time}`,
	).toISOString();
}

export function rangeLabelFor(range: DateRange): string {
	const { start, end } = range;
	if (!start) return 'Select dates';
	if (!end || (end.day === start.day && end.month === start.month && end.year === start.year))
		return fmtCalDate(start);
	return `${fmtCalDate(start)} – ${fmtCalDate(end)}`;
}

export function fmtDate(iso: string): string {
	return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

export function fmtDateTime(iso: string): string {
	return new Date(iso).toLocaleString(undefined, {
		month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
	});
}
