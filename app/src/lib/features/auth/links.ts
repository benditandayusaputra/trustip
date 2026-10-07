const signInPath = '/masuk';
const registerPath = '/daftar';
const defaultLanding = '/app';
const probeOrigin = 'https://trustip.invalid';

export function signInHref(next: string | null): string {
	return withNext(signInPath, next);
}

export function registerHref(next: string | null): string {
	return withNext(registerPath, next);
}

export function safeNextPath(raw: string | null): string {
	if (!raw?.startsWith('/') || raw.startsWith('//') || raw.startsWith('/\\')) {
		return defaultLanding;
	}
	const resolved = new URL(raw, probeOrigin);
	if (resolved.origin !== probeOrigin) {
		return defaultLanding;
	}
	return resolved.pathname + resolved.search + resolved.hash;
}

function withNext(path: string, next: string | null): string {
	return next ? `${path}?${new URLSearchParams({ lanjut: next }).toString()}` : path;
}
