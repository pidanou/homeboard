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

export function relativeDate(iso: string): string {
	const [y, m, d] = iso.slice(0, 10).split('-').map(Number);
	const targetMs = new Date(y, m - 1, d).getTime();
	const today = new Date();
	const todayMs = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime();
	const diff = Math.round((targetMs - todayMs) / 86400000);
	if (diff <= -2) return `${Math.abs(diff)} days ago`;
	if (diff === -1) return 'yesterday';
	if (diff === 0) return 'today';
	if (diff === 1) return 'tomorrow';
	if (diff < 7) return `in ${diff} days`;
	return fmtDate(iso);
}

export function localDayMs(iso: string): number {
	const [y, m, d] = iso.slice(0, 10).split('-').map(Number);
	return new Date(y, m - 1, d).getTime();
}

export function fmtDateTime(iso: string): string {
	return new Date(iso).toLocaleString(undefined, {
		month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
	});
}
