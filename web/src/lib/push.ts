const BASE = '/api/v1';

async function getVapidKey(): Promise<string> {
	const res = await fetch(`${BASE}/push/vapid-public-key`);
	const { public_key } = await res.json();
	return public_key;
}

export async function subscribePush(familyId: string): Promise<void> {
	if (!('serviceWorker' in navigator) || !('PushManager' in window)) return;

	const permission = await Notification.requestPermission();
	if (permission !== 'granted') return;

	const reg = await navigator.serviceWorker.ready;
	const vapidKey = await getVapidKey();

	const sub = await reg.pushManager.subscribe({
		userVisibleOnly: true,
		applicationServerKey: urlBase64ToUint8Array(vapidKey)
	});

	const { endpoint, keys } = sub.toJSON() as {
		endpoint: string;
		keys: { auth: string; p256dh: string };
	};

	await fetch(`${BASE}/households/${familyId}/push/subscribe`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ endpoint, auth: keys.auth, p256dh: keys.p256dh })
	});
}

export async function unsubscribePush(familyId: string): Promise<void> {
	if (!('serviceWorker' in navigator)) return;

	const reg = await navigator.serviceWorker.ready;
	const sub = await reg.pushManager.getSubscription();
	if (!sub) return;

	await fetch(`${BASE}/households/${familyId}/push/subscribe`, {
		method: 'DELETE',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ endpoint: sub.endpoint })
	});
	await sub.unsubscribe();
}

export async function isPushSubscribed(): Promise<boolean> {
	if (!('serviceWorker' in navigator)) return false;
	const reg = await navigator.serviceWorker.ready;
	return (await reg.pushManager.getSubscription()) !== null;
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
	const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
	const raw = atob(base64);
	return Uint8Array.from([...raw].map((c) => c.charCodeAt(0)));
}
