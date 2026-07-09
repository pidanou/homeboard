import { CalendarDate, type DateValue } from '@internationalized/date';
import type { DateRange } from 'bits-ui';
import { getLocale } from '$lib/paraglide/runtime';
import * as msg from '$lib/paraglide/messages.js';

export function calDateToISO(d: DateValue): string {
	const pad = (n: number) => String(n).padStart(2, '0');
	return `${d.year}-${pad(d.month)}-${pad(d.day)}T00:00:00.000Z`;
}

export function isoToCalDate(iso: string): CalendarDate {
	const [y, m, d] = iso.slice(0, 10).split('-').map(Number);
	return new CalendarDate(y, m, d);
}

export function fmtCalDate(d: DateValue): string {
	return new Date(
		`${d.year}-${String(d.month).padStart(2, '0')}-${String(d.day).padStart(2, '0')}`,
	).toLocaleDateString(getLocale(), { month: 'short', day: 'numeric', year: 'numeric' });
}

export function calDateTimeToISO(d: DateValue, time: string, allDay: boolean): string {
	const pad = (n: number) => String(n).padStart(2, '0');
	if (allDay) return calDateToISO(d);
	return new Date(`${d.year}-${pad(d.month)}-${pad(d.day)}T${time}`).toISOString();
}

export function rangeLabelFor(range: DateRange): string {
	const { start, end } = range;
	if (!start) return msg.dates_select_dates();
	if (!end || (end.day === start.day && end.month === start.month && end.year === start.year))
		return fmtCalDate(start);
	return `${fmtCalDate(start)} – ${fmtCalDate(end)}`;
}

export function fmtDate(iso: string): string {
	const [y, m, d] = iso.slice(0, 10).split('-').map(Number);
	return new Date(y, m - 1, d).toLocaleDateString(getLocale(), { month: 'short', day: 'numeric' });
}

export function relativeDate(iso: string): string {
	const [y, m, d] = iso.slice(0, 10).split('-').map(Number);
	const targetMs = new Date(y, m - 1, d).getTime();
	const today = new Date();
	const todayMs = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime();
	const diff = Math.round((targetMs - todayMs) / 86400000);
	if (diff <= -2) return msg.dates_days_ago({ days: Math.abs(diff) });
	if (diff === -1) return msg.dates_yesterday();
	if (diff === 0) return msg.dates_today();
	if (diff === 1) return msg.dates_tomorrow();
	if (diff < 7) return msg.dates_in_days({ days: diff });
	return fmtDate(iso);
}

export function localDayMs(iso: string): number {
	const [y, m, d] = iso.slice(0, 10).split('-').map(Number);
	return new Date(y, m - 1, d).getTime();
}

export function fmtTime(iso: string): string {
	return new Date(iso).toLocaleTimeString(getLocale(), { hour: '2-digit', minute: '2-digit' });
}

export function taskHasTime(iso: string): boolean {
	const d = new Date(iso);
	return d.getUTCHours() !== 0 || d.getUTCMinutes() !== 0;
}

export function isoToTimeInput(iso: string): string {
	const d = new Date(iso);
	return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}

export function fmtDateTime(iso: string): string {
	return new Date(iso).toLocaleString(getLocale(), {
		month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
	});
}

export function fmtWeekdayDate(d: Date): string {
	return d.toLocaleDateString(getLocale(), { weekday: 'long', month: 'long', day: 'numeric' });
}
