import type { SkeletonVariant } from '../ui/skeleton';

export type PageSkeletonKind = SkeletonVariant | 'store-dashboard' | 'consignor-dashboard';

const exact: Record<string, PageSkeletonKind> = {
	'/': 'home',
	'/masuk': 'auth',
	'/daftar': 'auth',
	'/demo': 'auth',
	'/onboarding': 'form',
	'/privasi': 'article',
	'/ketentuan': 'article',
	'/luring': 'detail',
	'/app': 'store-dashboard',
	'/app/toko': 'store-dashboard',
	'/app/penitip': 'consignor-dashboard',
	'/app/pengaturan': 'form',
	'/app/laporan': 'report',
	'/cari/toko': 'cards',
	'/cari/penitip': 'cards',
	'/kebutuhan': 'cards'
};

const patterns: [RegExp, PageSkeletonKind][] = [
	[/^\/(p|produk)\//u, 'product'],
	[/^\/(t|u)\//u, 'profile'],
	[/\/(baru|perubahan)$/u, 'form'],
	[/^\/app\/produk\/[^/]+$/u, 'form'],
	[/^\/(kebutuhan|verifikasi|undangan)\//u, 'detail'],
	[/^\/app\/[^/]+\//u, 'detail'],
	[/^\/app\//u, 'list']
];

export function pageSkeletonKind(path: string): PageSkeletonKind {
	const pathname = path.split(/[?#]/u)[0] ?? path;
	return exact[pathname] ?? patterns.find(([pattern]) => pattern.test(pathname))?.[1] ?? 'detail';
}

export function pendingPagePath(
	destination: Pick<URL, 'pathname'> | undefined,
	current: Pick<URL, 'pathname'>
): string | undefined {
	return destination?.pathname === current.pathname ? undefined : destination?.pathname;
}
