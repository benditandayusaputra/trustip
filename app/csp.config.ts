import type { Config } from '@sveltejs/kit/vite';

type CspDirectives = NonNullable<NonNullable<Config['csp']>['directives']>;
type Source = NonNullable<CspDirectives['img-src']>[number];

const captchaOrigin = 'https://challenges.cloudflare.com';
const mapTileOrigin = 'https://tile.openstreetmap.org';

interface PublicOrigins {
	files?: string | undefined;
	stream?: string | undefined;
}

const isHostOrigin = (origin: string): origin is Source & string =>
	/^https?:\/\/[a-z0-9.-]+(:\d+)?$/u.test(origin);

function originOf(url: string | undefined): Source[] {
	if (!url) {
		return [];
	}
	const { origin } = new URL(url);
	return isHostOrigin(origin) ? [origin] : [];
}

export function contentSecurityPolicy(origins: PublicOrigins): CspDirectives {
	return {
		'default-src': ['self'],
		'script-src': ['self', captchaOrigin],
		'style-src': ['self'],
		'style-src-attr': ['unsafe-inline'],
		'img-src': ['self', 'data:', 'blob:', mapTileOrigin, ...originOf(origins.files)],
		'font-src': ['self'],
		'connect-src': ['self', ...originOf(origins.stream)],
		'frame-src': [captchaOrigin],
		'worker-src': ['self'],
		'manifest-src': ['self'],
		'object-src': ['none'],
		'base-uri': ['self'],
		'form-action': ['self'],
		'frame-ancestors': ['none']
	};
}
