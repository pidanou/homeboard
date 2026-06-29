import { toast } from 'svelte-sonner';

export function getBaseUrl(): string {
	return (import.meta.env.PUBLIC_API_URL ?? 'http://localhost:8080').replace(/\/$/, '');
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;

	const res = await fetch(`${getBaseUrl()}${path}`, {
		...init,
		headers: {
			'Content-Type': 'application/json',
			...(token ? { Authorization: `Bearer ${token}` } : {}),
			...init.headers
		}
	}).catch((err) => {
		toast.error(`Cannot reach server: ${getBaseUrl()}`);
		throw err;
	});

	if (!res.ok) {
		const text = await res.text();
		toast.error(text || res.statusText);
		throw new Error(text || res.statusText);
	}

	if (res.status === 204) return null as T;
	return res.json() as Promise<T>;
}

export const api = {
	get: <T>(path: string) => request<T>(path),
	post: <T>(path: string, body: unknown) =>
		request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
	put: <T>(path: string, body: unknown) =>
		request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
	patch: <T>(path: string, body: unknown) =>
		request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),
	delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
	upload: async <T>(path: string, formData: FormData): Promise<T> => {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		const res = await fetch(`${getBaseUrl()}${path}`, {
			method: 'POST',
			body: formData,
			headers: token ? { Authorization: `Bearer ${token}` } : {}
		});
		if (!res.ok) {
			const text = (await res.text()) || res.statusText;
			toast.error(text);
			throw new Error(text);
		}
		if (res.status === 204) return null as T;
		return res.json() as Promise<T>;
	}
};

/** Returns an EventSource URL with the JWT token as a query param (EventSource can't set headers). */
export function sseUrl(path: string): string {
	const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
	const url = `${getBaseUrl()}${path}`;
	return token ? `${url}?token=${encodeURIComponent(token)}` : url;
}
