import { api } from '$lib/api/client';

async function getVapidKey(): Promise<string> {
	const result = await api.get<{ public_key: string }>('/api/v1/push/vapid-public-key');
	return result?.public_key ?? '';
}

export async function subscribePush(): Promise<void> {
	if (!('serviceWorker' in navigator) || !('PushManager' in window)) return;

	const permission = await Notification.requestPermission();
	if (permission !== 'granted') return;

	const reg = await navigator.serviceWorker.ready;

	// Clear any existing subscription — resubscribe fresh with current VAPID key
	const existing = await reg.pushManager.getSubscription();
	if (existing) await existing.unsubscribe();

	const vapidKey = await getVapidKey();

	const sub = await reg.pushManager.subscribe({
		userVisibleOnly: true,
		applicationServerKey: urlBase64ToUint8Array(vapidKey)
	});

	const endpoint = sub.endpoint;
	const p256dh = arrayBufferToBase64Url(sub.getKey('p256dh')!);
	const auth = arrayBufferToBase64Url(sub.getKey('auth')!);

	await api.post('/api/v1/push/subscribe', { endpoint, auth, p256dh });
}

export async function unsubscribePush(): Promise<void> {
	if (!('serviceWorker' in navigator)) return;

	const reg = await navigator.serviceWorker.ready;
	const sub = await reg.pushManager.getSubscription();
	if (!sub) return;

	await api.post('/api/v1/push/unsubscribe', { endpoint: sub.endpoint });
	await sub.unsubscribe();
}

export async function isPushSubscribed(): Promise<boolean> {
	if (!('serviceWorker' in navigator)) return false;
	const reg = await navigator.serviceWorker.ready;
	return (await reg.pushManager.getSubscription()) !== null;
}

function arrayBufferToBase64Url(buffer: ArrayBuffer): string {
	return btoa(String.fromCharCode(...new Uint8Array(buffer)))
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=/g, '');
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
	const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
	const raw = atob(base64);
	return Uint8Array.from([...raw].map((c) => c.charCodeAt(0)));
}
