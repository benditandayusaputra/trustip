import type { ApiClient } from '#lib/api/client.js';
import { listCities, listProvinces, type City, type Province } from '#lib/api/organizations.js';
import type { GeoPoint } from './schema';

export class RegionChoice {
	provinces = $state<Province[]>([]);
	cities = $state<City[]>([]);
	failed = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async loadProvinces(): Promise<void> {
		this.failed = false;
		try {
			this.provinces = await listProvinces(this.#client());
		} catch {
			this.failed = true;
		}
	}

	async loadCities(provinceCode: string): Promise<void> {
		this.cities = [];
		if (provinceCode === '') {
			return;
		}
		try {
			this.cities = await listCities(this.#client(), provinceCode);
		} catch {
			this.failed = true;
		}
	}

	centerOf(cityCode: string): GeoPoint | undefined {
		const city = this.cities.find((candidate) => candidate.code === cityCode);
		return city && { latitude: city.latitude, longitude: city.longitude };
	}
}
