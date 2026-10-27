import type { ApiClient } from '#lib/api/client.js';
import { ApiError, NetworkError } from '#lib/api/errors.js';
import { uploadImage } from '#lib/api/files.js';
import { getOrganization, updateProfile, type Organization } from '#lib/api/organizations.js';
import { organization as id } from '#lib/i18n/organization.js';
import { toOnboardingErrors, toProfileUpdateRequest, toProfileValues } from './mapper';
import { RegionChoice } from './regions.svelte';
import { emptyValues, hasErrors, validateProfile, type OnboardingErrors } from './schema';
import type { SubmitOutcome } from './wizard.svelte';

const firstServerErrorStatus = 500;

export type SettingsStatus = 'loading' | 'ready' | 'failed';
export type SaveOutcome = Exclude<SubmitOutcome, { kind: 'founded' }> | { kind: 'saved' };

export class ProfileSettings {
	status = $state<SettingsStatus>('loading');
	failure = $state('');
	organization = $state<Organization>();
	values = $state(emptyValues());
	logo = $state<Organization['logo']>(null);
	errors = $state<OnboardingErrors>({});
	logoError = $state('');
	busy = $state(false);
	isUploading = $state(false);
	readonly regions: RegionChoice;
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
		this.regions = new RegionChoice(client);
	}

	async load(organizationId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.#show(await getOrganization(this.#client(), organizationId));
			await Promise.all([
				this.regions.loadProvinces(),
				this.regions.loadCities(this.values.provinceCode)
			]);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async chooseLogo(file: File): Promise<void> {
		this.isUploading = true;
		this.logoError = '';
		try {
			const stored = await uploadImage(this.#client(), file, 'LOGO');
			this.logo = { id: stored.id, url: stored.url };
		} catch (error) {
			this.logoError = error instanceof ApiError ? error.message : id.settings.imageInvalid;
		} finally {
			this.isUploading = false;
		}
	}

	async save(): Promise<SaveOutcome> {
		const organization = this.organization;
		this.errors = validateProfile({ ...this.values, type: organization?.type ?? '' });
		if (!organization || hasErrors(this.errors)) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			const request = toProfileUpdateRequest(this.values, this.logo?.id ?? null);
			this.#show(await updateProfile(this.#client(), organization.id, request));
			return { kind: 'saved' };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	async reset(): Promise<void> {
		const shownProvince = this.values.provinceCode;
		if (this.organization) {
			this.#show(this.organization);
		}
		this.errors = {};
		this.logoError = '';
		if (this.values.provinceCode !== shownProvince) {
			await this.regions.loadCities(this.values.provinceCode);
		}
	}

	#show(organization: Organization): void {
		this.organization = organization;
		this.values = { ...emptyValues(), ...toProfileValues(organization), type: organization.type };
		this.logo = organization.logo;
	}

	#explain(error: unknown): SaveOutcome {
		if (error instanceof NetworkError) {
			return { kind: 'failed', message: error.message, canRetry: true };
		}
		if (!(error instanceof ApiError)) {
			throw error;
		}
		if (error.fields.length > 0) {
			this.errors = toOnboardingErrors(error.fields);
			return { kind: 'invalid' };
		}
		return {
			kind: 'failed',
			message: error.message,
			canRetry: error.status >= firstServerErrorStatus
		};
	}
}
