export type PlaceKind = 'ORGANIZATION' | 'HERE' | 'CITY';

export interface Place {
	kind: PlaceKind;
	lat: string;
	lng: string;
	cityCode: string;
}
