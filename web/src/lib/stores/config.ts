import { writable } from 'svelte/store';
import { api } from '$lib/api/client';

export const allowMultiHousehold = writable(false);

export async function loadConfig() {
	try {
		const config = await api.get<{ allowMultiHousehold: boolean }>('/api/v1/config');
		allowMultiHousehold.set(config?.allowMultiHousehold ?? false);
	} catch {
		allowMultiHousehold.set(false);
	}
}
