import { api } from '$lib/api/client';

export type CurrentUser = {
	id: string;
	name: string;
	email: string;
	avatar_url: string | null;
	locale: string;
	time_format: 'auto' | '12' | '24';
};

let user = $state<CurrentUser | null>(null);

export const currentUser = {
	get value() {
		return user;
	},
};

export async function loadCurrentUser() {
	try {
		user = await api.get<CurrentUser>('/api/v1/profile');
	} catch {
		user = null;
	}
}
