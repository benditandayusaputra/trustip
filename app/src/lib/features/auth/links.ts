const signInPath = '/masuk';

export function signInHref(next: string | null): string {
	return next ? `${signInPath}?${new URLSearchParams({ lanjut: next }).toString()}` : signInPath;
}
