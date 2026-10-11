const dashboards: Record<string, string> = {
	CONSIGNOR: '/app/penitip',
	STORE: '/app/toko'
};

export function dashboardDestination(path: string, organizationType: string): string {
	const home = dashboards[organizationType] ?? '/app';
	const pathname = path.split(/[?#]/u)[0] ?? path;
	const isDashboard = Object.values(dashboards).some(
		(dashboard) => pathname === dashboard || pathname.startsWith(`${dashboard}/`)
	);
	if (pathname === '/app' || (isDashboard && pathname !== home)) {
		return home;
	}
	return path;
}
