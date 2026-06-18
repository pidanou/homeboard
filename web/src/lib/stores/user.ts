import { writable } from 'svelte/store';
import { api } from '$lib/api/client';

export type CurrentUser = {
	id: string;
	name: string;
	email: string;
	avatar_url: string | null;
};

export const currentUser = writable<CurrentUser | null>(null);

export async function loadCurrentUser() {
	try {
		const user = await api.get<CurrentUser>('/api/v1/profile');
		currentUser.set(user);
	} catch {
		currentUser.set(null);
	}
}
