/// <reference lib="webworker" />
import { cleanupOutdatedCaches, precacheAndRoute } from 'workbox-precaching';

declare let self: ServiceWorkerGlobalScope;

cleanupOutdatedCaches();
precacheAndRoute(self.__WB_MANIFEST);

self.addEventListener('push', (event) => {
	const data = event.data?.json() ?? {};
	const title: string = data.title ?? 'Homeboard';
	const body: string = data.body ?? '';
	event.waitUntil(
		self.registration.showNotification(title, {
			body,
			icon: '/icon-192.png',
			badge: '/icon-192.png'
		})
	);
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	event.waitUntil(clients.openWindow('/'));
});
