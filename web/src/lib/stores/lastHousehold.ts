import { browser } from '$app/environment';

const KEY = 'last-household-id';

export function getLastHouseholdId(): string | null {
	if (!browser) return null;
	return localStorage.getItem(KEY);
}

export function setLastHouseholdId(id: string): void {
	if (!browser) return;
	localStorage.setItem(KEY, id);
}

export function clearLastHouseholdId(): void {
	if (!browser) return;
	localStorage.removeItem(KEY);
}
