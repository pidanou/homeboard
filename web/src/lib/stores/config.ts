import { writable } from 'svelte/store';
import { api } from '$lib/api/client';

export const allowMultiHousehold = writable(false);
export const oidcEnabled = writable(false);
export const oidcProviderName = writable('SSO');
export const allowPasswordLogin = writable(true);

interface AppConfig {
	allowMultiHousehold: boolean;
	oidcEnabled: boolean;
	oidcProviderName: string;
	allowPasswordLogin: boolean;
}

export async function loadConfig() {
	try {
		const config = await api.get<AppConfig>('/api/v1/config');
		allowMultiHousehold.set(config?.allowMultiHousehold ?? false);
		oidcEnabled.set(config?.oidcEnabled ?? false);
		oidcProviderName.set(config?.oidcProviderName || 'SSO');
		allowPasswordLogin.set(config?.allowPasswordLogin ?? true);
	} catch {
		allowMultiHousehold.set(false);
		oidcEnabled.set(false);
		oidcProviderName.set('SSO');
		allowPasswordLogin.set(true);
	}
}
