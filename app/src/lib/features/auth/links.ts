import type { Account } from '#lib/api/users.js';
import { dashboardDestination } from '#lib/domain/dashboard-access.js';

const signInPath = '/masuk';
const registerPath = '/daftar';
const onboardingPath = '/onboarding';
const defaultLanding = '/app';
const probeOrigin = 'https://trustip.invalid';

export function signInHref(next: string | null): string {
	return withNext(signInPath, next);
}

export function registerHref(next: string | null): string {
	return withNext(registerPath, next);
}

export function onboardingHref(next: string | null): string {
	return withNext(onboardingPath, next);
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

export function signedInPath(next: string | null, account: Account): string {
	const path = safeNextPath(next);
	if (!account.membership) {
		return onboardingHref(path);
	}
	return dashboardDestination(path, account.membership.organization_type);
}
