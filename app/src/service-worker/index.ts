import { self } from '$app/service-worker';
import { assets, immutable, prerendered } from '$app/manifest';
import { version } from '$app/env';

const shellCache = `trustip-shell-${version}`;
const pageCache = `trustip-pages-${version}`;
const keptCaches = new Set([shellCache, pageCache]);
const offlinePage = '/luring';
const shellFiles = [...immutable, ...assets, ...prerendered].map(
	(file) => new URL(file.path, self.registration.scope).pathname
);

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
				Promise.all(keys.filter((key) => !keptCaches.has(key)).map((key) => caches.delete(key)))
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
		event.respondWith(networkOrSavedPage(request, url));
	}
});

async function cachedFirst(request: Request): Promise<Response> {
	const cached = await caches.match(request);
	return cached ?? fetch(request);
}

async function networkOrSavedPage(request: Request, url: URL): Promise<Response> {
	try {
		const response = await fetch(request);
		if (response.ok && url.pathname.startsWith('/app')) {
			const cache = await caches.open(pageCache);
			await cache.put(request, response.clone());
		}
		return response;
	} catch {
		const saved = await caches.match(request);
		return saved ?? Response.redirect(new URL(offlinePage, self.location.href).href);
	}
}
