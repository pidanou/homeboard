import { writable } from 'svelte/store';

export type HouseholdEntry = { id: string; name: string };

export const households = writable<HouseholdEntry[]>([]);

export function updateHouseholdName(id: string, name: string) {
	households.update(hs => hs.map(h => (h.id === id ? { ...h, name } : h)));
}
