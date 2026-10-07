const probeOrigin = 'https://trustip.invalid';

export function safeHref(raw: string): string | undefined {
	if (raw.startsWith('/') && !raw.startsWith('//') && !raw.startsWith('/\\')) {
		const resolved = new URL(raw, probeOrigin);
		return resolved.origin === probeOrigin ? resolved.pathname + resolved.search : undefined;
	}
	try {
		const url = new URL(raw);
		return url.protocol === 'https:' ? url.href : undefined;
	} catch {
		return undefined;
	}
}
