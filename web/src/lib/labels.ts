export type LabelColor =
	| 'red' | 'orange' | 'yellow' | 'green'
	| 'teal' | 'blue' | 'purple' | 'pink' | 'gray';

export const LABEL_COLORS: LabelColor[] = [
	'red', 'orange', 'yellow', 'green', 'teal', 'blue', 'purple', 'pink', 'gray',
];

export const LABEL_CHIP: Record<LabelColor, string> = {
	red:    'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
	orange: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
	yellow: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
	green:  'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
	teal:   'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400',
	blue:   'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
	purple: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
	pink:   'bg-pink-100 text-pink-700 dark:bg-pink-900/30 dark:text-pink-400',
	gray:   'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
};

export const LABEL_DOT: Record<LabelColor, string> = {
	red: 'bg-red-500', orange: 'bg-orange-500', yellow: 'bg-yellow-500',
	green: 'bg-green-500', teal: 'bg-teal-500', blue: 'bg-blue-500',
	purple: 'bg-purple-500', pink: 'bg-pink-500', gray: 'bg-gray-400',
};

export function chipClass(color: string): string {
	return LABEL_CHIP[color as LabelColor] ?? LABEL_CHIP.gray;
}

export function dotClass(color: string): string {
	return LABEL_DOT[color as LabelColor] ?? LABEL_DOT.gray;
}
