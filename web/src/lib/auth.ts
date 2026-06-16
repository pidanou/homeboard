import { browser } from '$app/environment';
import { goto } from '$app/navigation';

export function getToken(): string | null {
	return browser ? localStorage.getItem('token') : null;
}

export function setToken(token: string): void {
	localStorage.setItem('token', token);
}

export function clearToken(): void {
	localStorage.removeItem('token');
}

export function logout(): void {
	clearToken();
	goto('/login');
}

export function isLoggedIn(): boolean {
	return !!getToken();
}
