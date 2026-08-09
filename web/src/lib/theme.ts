import { writable } from 'svelte/store';

export const isDark = writable(false);

export function initTheme() {
	isDark.set(document.documentElement.classList.contains('dark'));
}

export function toggleTheme() {
	const dark = document.documentElement.classList.toggle('dark');
	localStorage.setItem('theme', dark ? 'dark' : 'light');
	isDark.set(dark);
}

export const DEFAULT_ACCENT = '#c2703c';

export const accentColor = writable<string>(DEFAULT_ACCENT);

function relativeLuminance(hex: string) {
	const c = hex.replace('#', '');
	const lin = (v: number) => (v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4));
	return (
		0.2126 * lin(parseInt(c.substring(0, 2), 16) / 255) +
		0.7152 * lin(parseInt(c.substring(2, 4), 16) / 255) +
		0.0722 * lin(parseInt(c.substring(4, 6), 16) / 255)
	);
}

export function initAccentColor() {
	accentColor.set(localStorage.getItem('accentColor') ?? DEFAULT_ACCENT);
}

export function setAccentColor(hex: string) {
	const foreground = relativeLuminance(hex) > 0.45 ? '#1a1410' : '#fdfaf7';
	const root = document.documentElement.style;
	root.setProperty('--primary', hex);
	root.setProperty('--primary-foreground', foreground);
	root.setProperty('--ring', hex);
	root.setProperty('--sidebar-primary', hex);
	root.setProperty('--sidebar-primary-foreground', foreground);
	root.setProperty('--sidebar-ring', hex);
	localStorage.setItem('accentColor', hex);
	accentColor.set(hex);
}

export function resetAccentColor() {
	const root = document.documentElement.style;
	for (const prop of [
		'--primary',
		'--primary-foreground',
		'--ring',
		'--sidebar-primary',
		'--sidebar-primary-foreground',
		'--sidebar-ring',
	]) {
		root.removeProperty(prop);
	}
	localStorage.removeItem('accentColor');
	accentColor.set(DEFAULT_ACCENT);
}
