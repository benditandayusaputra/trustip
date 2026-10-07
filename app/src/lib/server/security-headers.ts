const securityHeaders: Record<string, string> = {
	'Strict-Transport-Security': 'max-age=63072000; includeSubDomains',
	'X-Content-Type-Options': 'nosniff',
	'Referrer-Policy': 'strict-origin-when-cross-origin',
	'Permissions-Policy': 'geolocation=(self), camera=(), microphone=(), payment=()',
	'Cross-Origin-Opener-Policy': 'same-origin',
	'X-Frame-Options': 'DENY',
	'Cross-Origin-Resource-Policy': 'same-origin'
};

export function applySecurityHeaders(headers: Headers): void {
	Object.entries(securityHeaders).forEach(([name, value]) => {
		headers.set(name, value);
	});
}
