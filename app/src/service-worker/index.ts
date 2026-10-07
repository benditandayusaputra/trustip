import { self } from '$app/service-worker';
import { assets, immutable, prerendered } from '$app/manifest';
import { version } from '$app/env';

const shellCache = `trustip-shell-${version}`;
const offlinePage = '/luring';
const shellFiles = [...immutable, ...assets, ...prerendered].map((file) => file.path);

self.addEventListener('install', (event) => {
	event.waitUntil(
		caches
			.open(shellCache)
			.then((cache) => cache.addAll(shellFiles))
			.then(() => self.skipWaiting())
	);
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(keys.filter((key) => key !== shellCache).map((key) => caches.delete(key)))
			)
			.then(() => self.clients.claim())
	);
});

self.addEventListener('fetch', (event) => {
	const { request } = event;
	const url = new URL(request.url);
	if (
		request.method !== 'GET' ||
		url.origin !== self.location.origin ||
		url.pathname.startsWith('/api/')
	) {
		return;
	}
	if (shellFiles.includes(url.pathname)) {
		event.respondWith(cachedFirst(request));
		return;
	}
	if (request.mode === 'navigate') {
		event.respondWith(networkOrOfflinePage(request));
	}
});

async function cachedFirst(request: Request): Promise<Response> {
	const cached = await caches.match(request);
	return cached ?? fetch(request);
}

async function networkOrOfflinePage(request: Request): Promise<Response> {
	try {
		return await fetch(request);
	} catch {
		const offline = await caches.match(offlinePage);
		return offline ?? Response.error();
	}
}
