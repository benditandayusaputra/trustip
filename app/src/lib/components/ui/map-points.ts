export interface MapMarker {
	id: string;
	label: string;
	latitude: number;
	longitude: number;
}

export interface MarkerGroup {
	point: MapMarker;
	members: MapMarker[];
}

export function markerGroups(points: MapMarker[]): MarkerGroup[] {
	const grouped = new Map<string, MarkerGroup>();
	for (const point of points) {
		const key = `${String(point.latitude)},${String(point.longitude)}`;
		const group = grouped.get(key);
		if (group) group.members.push(point);
		else grouped.set(key, { point, members: [point] });
	}
	return [...grouped.values()];
}
