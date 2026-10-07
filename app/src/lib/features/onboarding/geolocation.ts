import type { GeoPoint } from './schema';

const locateTimeoutMs = 10_000;

export function locateDevice(): Promise<GeoPoint> {
	return new Promise((resolve, reject) => {
		navigator.geolocation.getCurrentPosition(
			(position) => {
				resolve({ latitude: position.coords.latitude, longitude: position.coords.longitude });
			},
			(failure) => {
				reject(new Error(failure.message));
			},
			{ enableHighAccuracy: true, timeout: locateTimeoutMs }
		);
	});
}
