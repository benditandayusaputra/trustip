import type { ApiClient } from '#lib/api/client.js';
import { updateConsignorProfile, type ConsignorProfile } from '#lib/api/organizations.js';
import { explainSectionFailure, type SectionOutcome } from './section';

type ConsignorField =
	| 'deliveryRadius'
	| 'deliveryDays'
	| 'preferredStoreTypes'
	| 'businessIdNo'
	| 'distributionPermitNo'
	| 'halalCertNo';

const fieldsFromServer: Record<string, ConsignorField | undefined> = {
	delivery_radius_km: 'deliveryRadius',
	delivery_days: 'deliveryDays',
	preferred_store_types: 'preferredStoreTypes',
	business_id_no: 'businessIdNo',
	distribution_permit_no: 'distributionPermitNo',
	halal_cert_no: 'halalCertNo'
};

export interface ConsignorValues {
	deliveryRadiusKm: number;
	deliveryDays: string[];
	preferredStoreTypes: string[];
	businessIdNo: string;
	distributionPermitNo: string;
	halalCertNo: string;
	canProvideSample: boolean;
}

const emptyProfile: ConsignorProfile = {
	delivery_radius_km: null,
	delivery_days: [],
	preferred_store_types: [],
	business_id_no: '',
	distribution_permit_no: '',
	halal_cert_no: '',
	can_provide_sample: false
};

function toValues(stored: ConsignorProfile | null): ConsignorValues {
	const profile = stored ?? emptyProfile;
	return {
		deliveryRadiusKm: profile.delivery_radius_km ?? 0,
		deliveryDays: profile.delivery_days,
		preferredStoreTypes: profile.preferred_store_types,
		businessIdNo: profile.business_id_no,
		distributionPermitNo: profile.distribution_permit_no,
		halalCertNo: profile.halal_cert_no,
		canProvideSample: profile.can_provide_sample
	};
}

function toRequest(values: ConsignorValues): ConsignorProfile {
	return {
		delivery_radius_km: values.deliveryRadiusKm > 0 ? values.deliveryRadiusKm : null,
		delivery_days: values.deliveryDays as ConsignorProfile['delivery_days'],
		preferred_store_types: values.preferredStoreTypes as ConsignorProfile['preferred_store_types'],
		business_id_no: values.businessIdNo,
		distribution_permit_no: values.distributionPermitNo,
		halal_cert_no: values.halalCertNo,
		can_provide_sample: values.canProvideSample
	};
}

export class ConsignorProfileEditor {
	values = $state<ConsignorValues>(toValues(null));
	errors = $state<Partial<Record<ConsignorField, string>>>({});
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient, profile: ConsignorProfile | null) {
		this.#client = client;
		this.values = toValues(profile);
	}

	async save(organizationId: string): Promise<SectionOutcome> {
		this.busy = true;
		this.errors = {};
		try {
			const saved = await updateConsignorProfile(
				this.#client(),
				organizationId,
				toRequest(this.values)
			);
			return { kind: 'saved', organization: saved };
		} catch (error) {
			const failure = explainSectionFailure(error, fieldsFromServer);
			this.errors = failure.errors;
			return failure.outcome;
		} finally {
			this.busy = false;
		}
	}
}
