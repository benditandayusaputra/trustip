const actionKey = 'trustip-first-action';
const dismissKey = 'trustip-install-dismissed';
export const actionDoneEvent = 'trustip:action-done';
const marked = '1';
const sessionFlags = new Set<string>();

function readFlag(key: string): boolean {
	try {
		return localStorage.getItem(key) === marked || sessionFlags.has(key);
	} catch {
		return sessionFlags.has(key);
	}
}

function writeFlag(key: string): void {
	try {
		localStorage.setItem(key, marked);
	} catch {
		sessionFlags.add(key);
	}
}

export function noteCompletedAction(): void {
	writeFlag(actionKey);
	window.dispatchEvent(new Event(actionDoneEvent));
}

export function hasCompletedAction(): boolean {
	return readFlag(actionKey);
}

export function dismissInstall(): void {
	writeFlag(dismissKey);
}

export function isInstallDismissed(): boolean {
	return readFlag(dismissKey);
}
