export type LabelColor =
	| 'red' | 'orange' | 'yellow' | 'green'
	| 'teal' | 'blue' | 'purple' | 'pink' | 'gray';

export const LABEL_COLORS: LabelColor[] = [
	'red', 'orange', 'yellow', 'green', 'teal', 'blue', 'purple', 'pink', 'gray',
];

// Legacy named colors map to a hex value; custom colors (from the color wheel) are
// already stored as hex and pass through unchanged.
export const CATEGORY_HEX: Record<string, string> = {
	red: '#f43f5e', orange: '#fb923c', yellow: '#f59e0b', green: '#059669',
	teal: '#0d9488', blue: '#6366f1', purple: '#8b5cf6', pink: '#f472b6', gray: '#a8a29e',
};

export function resolveHex(color: string): string {
	return color.startsWith('#') ? color : (CATEGORY_HEX[color] ?? CATEGORY_HEX.gray);
}

export function dotStyle(color: string): string {
	return `background-color:${resolveHex(color)}`;
}

export function chipStyle(color: string): string {
	const hex = resolveHex(color);
	return `background-color:color-mix(in srgb, ${hex} 16%, transparent);color:${hex}`;
}
