import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { uploadImage } from '#lib/api/files.js';
import {
	listCategories,
	updateStoreProfile,
	type Category,
	type StoreProfile,
	type StoreProfileRequest,
	type StoreType
} from '#lib/api/organizations.js';
import { id } from '#lib/i18n/id.js';
import { explainSectionFailure, type SectionOutcome } from './section';

export const maxStorePhotos = 5;
const basisPointsPerPercent = 100;
const maxCommissionPercent = 50;

type Photo = StoreProfile['photos'][number];

export interface StoreProfileValues {
	storeType: StoreType | '';
	categoryIds: string[];
	acceptsFixedDeposit: boolean;
	acceptsCommission: boolean;
	commissionMinPercent: number;
	commissionMaxPercent: number;
	requiresDistributionPermit: boolean;
	requiresHalal: boolean;
	requiresSample: boolean;
	otherRequirements: string;
	openingHours: string;
	photos: Photo[];
}

type StoreProfileField = 'storeType' | 'categoryIds' | 'commissionMax' | 'photos' | 'text';
export type StoreProfileErrors = Partial<Record<StoreProfileField, string>>;

const fieldsFromServer: Record<string, StoreProfileField | undefined> = {
	store_type: 'storeType',
	accepted_category_ids: 'categoryIds',
	commission_max_bps: 'commissionMax',
	photo_file_ids: 'photos',
	other_requirements: 'text',
	opening_hours: 'text'
};

const emptyStoreProfile: StoreProfile = {
	store_type: null,
	accepted_category_ids: [],
	accepts_fixed_deposit: false,
	accepts_commission: false,
	commission_min_bps: null,
	commission_max_bps: null,
	requires_distribution_permit: false,
	requires_halal: false,
	requires_sample: false,
	other_requirements: '',
	opening_hours: '',
	photos: []
};

function toStoreProfileValues(stored: StoreProfile | null): StoreProfileValues {
	const profile = stored ?? emptyStoreProfile;
	return {
		storeType: profile.store_type ?? '',
		categoryIds: profile.accepted_category_ids,
		acceptsFixedDeposit: profile.accepts_fixed_deposit,
		acceptsCommission: profile.accepts_commission,
		commissionMinPercent: percentOf(profile.commission_min_bps),
		commissionMaxPercent: percentOf(profile.commission_max_bps),
		requiresDistributionPermit: profile.requires_distribution_permit,
		requiresHalal: profile.requires_halal,
		requiresSample: profile.requires_sample,
		otherRequirements: profile.other_requirements,
		openingHours: profile.opening_hours,
		photos: profile.photos
	};
}

function percentOf(basisPoints: number | null): number {
	return (basisPoints ?? 0) / basisPointsPerPercent;
}

function toStoreProfileRequest(values: StoreProfileValues): StoreProfileRequest {
	const commission = values.acceptsCommission;
	return {
		store_type: values.storeType === '' ? null : values.storeType,
		accepted_category_ids: values.categoryIds,
		accepts_fixed_deposit: values.acceptsFixedDeposit,
		accepts_commission: commission,
		commission_min_bps: commission ? values.commissionMinPercent * basisPointsPerPercent : null,
		commission_max_bps: commission ? values.commissionMaxPercent * basisPointsPerPercent : null,
		requires_distribution_permit: values.requiresDistributionPermit,
		requires_halal: values.requiresHalal,
		requires_sample: values.requiresSample,
		other_requirements: values.otherRequirements,
		opening_hours: values.openingHours,
		photo_file_ids: values.photos.map((photo) => photo.id)
	};
}

function validateStoreProfile(values: StoreProfileValues): StoreProfileErrors {
	const isOrdered = values.commissionMinPercent <= values.commissionMaxPercent;
	if (
		values.acceptsCommission &&
		(!isOrdered || values.commissionMaxPercent > maxCommissionPercent)
	) {
		return { commissionMax: id.storeProfile.commissionOrder };
	}
	return {};
}

export class StoreProfileEditor {
	values = $state<StoreProfileValues>(toStoreProfileValues(null));
	categories = $state<Category[]>([]);
	errors = $state<StoreProfileErrors>({});
	busy = $state(false);
	isUploading = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient, profile: StoreProfile | null) {
		this.#client = client;
		this.values = toStoreProfileValues(profile);
	}

	async loadCategories(): Promise<void> {
		this.categories = await listCategories(this.#client()).catch(() => []);
	}

	async addPhoto(file: File): Promise<void> {
		this.isUploading = true;
		this.errors = { ...this.errors, photos: '' };
		try {
			const stored = await uploadImage(this.#client(), file, 'STORE_PHOTO');
			this.values.photos = [...this.values.photos, { id: stored.id, url: stored.url }];
		} catch (error) {
			this.errors.photos = error instanceof ApiError ? error.message : id.settings.imageInvalid;
		} finally {
			this.isUploading = false;
		}
	}

	removePhoto(photoId: string): void {
		this.values.photos = this.values.photos.filter((photo) => photo.id !== photoId);
	}

	async save(organizationId: string): Promise<SectionOutcome> {
		this.errors = validateStoreProfile(this.values);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			const request = toStoreProfileRequest(this.values);
			return {
				kind: 'saved',
				organization: await updateStoreProfile(this.#client(), organizationId, request)
			};
		} catch (error) {
			const failure = explainSectionFailure(error, fieldsFromServer);
			this.errors = failure.errors;
			return failure.outcome;
		} finally {
			this.busy = false;
		}
	}
}
