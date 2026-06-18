export type LabelColor =
	| 'red' | 'orange' | 'yellow' | 'green'
	| 'teal' | 'blue' | 'purple' | 'pink' | 'gray';

export const LABEL_COLORS: LabelColor[] = [
	'red', 'orange', 'yellow', 'green', 'teal', 'blue', 'purple', 'pink', 'gray',
];

export const CATEGORY_HEX: Record<string, string> = {
	red: '#f43f5e', orange: '#fb923c', yellow: '#f59e0b', green: '#059669',
	teal: '#0d9488', blue: '#6366f1', purple: '#8b5cf6', pink: '#f472b6', gray: '#a8a29e',
};

const LABEL_CHIP: Record<LabelColor, string> = {
	red:    'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400',
	orange: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
	yellow: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-500',
	green:  'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
	teal:   'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400',
	blue:   'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400',
	purple: 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400',
	pink:   'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-300',
	gray:   'bg-stone-100 text-stone-600 dark:bg-stone-800 dark:text-stone-300',
};

const LABEL_DOT: Record<LabelColor, string> = {
	red: 'bg-rose-500', orange: 'bg-orange-400', yellow: 'bg-amber-400',
	green: 'bg-emerald-600', teal: 'bg-teal-600', blue: 'bg-indigo-500',
	purple: 'bg-violet-500', pink: 'bg-pink-400', gray: 'bg-stone-400',
};

export function chipClass(color: string): string {
	return LABEL_CHIP[color as LabelColor] ?? LABEL_CHIP.gray;
}

export function dotClass(color: string): string {
	return LABEL_DOT[color as LabelColor] ?? LABEL_DOT.gray;
}
