import type { ApiClient } from '#lib/api/client.js';
import { ApiError, NetworkError } from '#lib/api/errors.js';
import {
	createOrganization,
	listCities,
	listProvinces,
	type City,
	type Province
} from '#lib/api/organizations.js';
import { refreshSession } from '#lib/api/sessions.js';
import { toOnboardingErrors, toOnboardingRequest } from './mapper';
import {
	emptyValues,
	hasErrors,
	stepOf,
	steps,
	validateAll,
	validateStep,
	type OnboardingErrors,
	type GeoPoint,
	type OnboardingField,
	type Step
} from './schema';

const firstServerErrorStatus = 500;

export type SubmitOutcome =
	| { kind: 'founded' }
	| { kind: 'invalid' }
	| { kind: 'failed'; message: string; canRetry: boolean };

export class OnboardingWizard {
	step = $state<Step>('type');
	values = $state(emptyValues());
	errors = $state<OnboardingErrors>({});
	provinces = $state<Province[]>([]);
	cities = $state<City[]>([]);
	regionsFailed = $state(false);
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	get mapCenter(): GeoPoint | undefined {
		const city = this.cities.find((candidate) => candidate.code === this.values.cityCode);
		return this.values.location ?? (city && { latitude: city.latitude, longitude: city.longitude });
	}

	pinAt(point: GeoPoint): void {
		this.values.location = point;
	}

	chooseCity(): void {
		this.values.location = null;
	}

	get stepNumber(): number {
		return steps.indexOf(this.step) + 1;
	}

	next(): boolean {
		this.errors = validateStep(this.step, this.values);
		const following = steps[this.stepNumber];
		if (hasErrors(this.errors) || following === undefined) {
			return false;
		}
		this.step = following;
		return true;
	}

	back(): void {
		this.errors = {};
		this.step = steps[steps.indexOf(this.step) - 1] ?? 'type';
	}

	async loadProvinces(): Promise<void> {
		this.regionsFailed = false;
		try {
			this.provinces = await listProvinces(this.#client());
		} catch {
			this.regionsFailed = true;
		}
	}

	async retryRegions(): Promise<void> {
		await this.loadProvinces();
		await this.chooseProvince(this.values.provinceCode);
	}

	async chooseProvince(provinceCode: string): Promise<void> {
		this.values.cityCode = '';
		this.values.location = null;
		this.cities = [];
		if (provinceCode === '') {
			return;
		}
		try {
			this.cities = await listCities(this.#client(), provinceCode);
		} catch {
			this.regionsFailed = true;
		}
	}

	async submit(): Promise<SubmitOutcome> {
		this.errors = validateAll(this.values);
		if (hasErrors(this.errors)) {
			return this.#showFirstProblem();
		}
		this.busy = true;
		try {
			await createOrganization(this.#client(), toOnboardingRequest(this.values));
			await refreshSession(this.#client());
			return { kind: 'founded' };
		} catch (error) {
			return await this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	#showFirstProblem(): SubmitOutcome {
		const first = Object.keys(this.errors)[0] as OnboardingField | undefined;
		this.step = first ? stepOf(first) : this.step;
		return { kind: 'invalid' };
	}

	async #explain(error: unknown): Promise<SubmitOutcome> {
		if (error instanceof NetworkError) {
			return { kind: 'failed', message: error.message, canRetry: true };
		}
		if (!(error instanceof ApiError)) {
			throw error;
		}
		if (error.fields.length > 0) {
			this.errors = toOnboardingErrors(error.fields);
			return this.#showFirstProblem();
		}
		if (error.code === 'ORGANIZATION_EXISTS') {
			await refreshSession(this.#client());
			return { kind: 'founded' };
		}
		return {
			kind: 'failed',
			message: error.message,
			canRetry: error.status >= firstServerErrorStatus
		};
	}
}
