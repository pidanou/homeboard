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
