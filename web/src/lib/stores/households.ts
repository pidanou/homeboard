import { writable } from 'svelte/store';

export type HouseholdEntry = { id: string; name: string; photo_url?: string | null; photo_original_url?: string | null; wallpaper_url?: string | null; wallpaper_original_url?: string | null };

export const households = writable<HouseholdEntry[]>([]);

export function updateHouseholdName(id: string | undefined, name: string) {
	if (!id) return;
	households.update(hs => hs.map(h => (h.id === id ? { ...h, name } : h)));
}

export function updateHouseholdPhoto(id: string | undefined, photo_url: string | null) {
	if (!id) return;
	households.update(hs => hs.map(h => (h.id === id ? { ...h, photo_url } : h)));
}

export function updateHouseholdWallpaper(id: string | undefined, wallpaper_url: string | null) {
	if (!id) return;
	households.update(hs => hs.map(h => (h.id === id ? { ...h, wallpaper_url } : h)));
}
