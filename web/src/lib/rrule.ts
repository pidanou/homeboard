import { CalendarDate } from '@internationalized/date';

export type RepeatFreq = 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
export type RepeatEndType = 'never' | 'until' | 'count';
export type RepeatEnd =
	| { type: 'never' }
	| { type: 'until'; date: CalendarDate }
	| { type: 'count'; count: number };

const FREQ: Record<Exclude<RepeatFreq, 'none'>, string> = {
	daily: 'DAILY', weekly: 'WEEKLY', monthly: 'MONTHLY', yearly: 'YEARLY',
};
const FREQ_REVERSE: Record<string, RepeatFreq> = {
	DAILY: 'daily', WEEKLY: 'weekly', MONTHLY: 'monthly', YEARLY: 'yearly',
};

// UNTIL is set to end-of-day UTC so the chosen date stays included regardless
// of the event's time-of-day (RFC5545 requires UNTIL to match DTSTART's value type).
export function buildRRule(freq: RepeatFreq, end: RepeatEnd): string | undefined {
	if (freq === 'none') return undefined;
	const pad = (n: number) => String(n).padStart(2, '0');
	let rule = `FREQ=${FREQ[freq]}`;
	if (end.type === 'until') {
		rule += `;UNTIL=${end.date.year}${pad(end.date.month)}${pad(end.date.day)}T235959Z`;
	} else if (end.type === 'count' && end.count > 0) {
		rule += `;COUNT=${end.count}`;
	}
	return rule;
}

export function parseRRule(rule: string | null | undefined): { freq: RepeatFreq; end: RepeatEnd } {
	if (!rule) return { freq: 'none', end: { type: 'never' } };
	const parts: Record<string, string> = {};
	for (const part of rule.replace(/^RRULE:/, '').split(';')) {
		const [key, value] = part.split('=');
		if (key && value) parts[key] = value;
	}
	const freq = FREQ_REVERSE[parts.FREQ] ?? 'none';
	if (parts.UNTIL) {
		const m = /^(\d{4})(\d{2})(\d{2})/.exec(parts.UNTIL);
		if (m) return { freq, end: { type: 'until', date: new CalendarDate(+m[1], +m[2], +m[3]) } };
	}
	if (parts.COUNT) return { freq, end: { type: 'count', count: Number(parts.COUNT) } };
	return { freq, end: { type: 'never' } };
}
