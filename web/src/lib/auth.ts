import { browser } from '$app/environment';
import { goto } from '$app/navigation';

export function getToken(): string | null {
	if (!browser) return null;
	return localStorage.getItem('token') ?? sessionStorage.getItem('token');
}

export function setToken(token: string, remember = true): void {
	if (remember) {
		localStorage.setItem('token', token);
	} else {
		sessionStorage.setItem('token', token);
	}
}

export function clearToken(): void {
	localStorage.removeItem('token');
	sessionStorage.removeItem('token');
}

export function logout(): void {
	clearToken();
	goto('/login');
}

export function isLoggedIn(): boolean {
	return !!getToken();
}
