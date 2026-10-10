import type { NeedRequest } from '#lib/api/needs.js';
import { needs as id } from '#lib/i18n/needs.js';

const minTitleLength = 5;
const maxTitleLength = 80;
const basisPointsPerPercent = 100;

export interface NeedValues {
	title: string;
	categoryId: string;
	description: string;
	weeklyQuantity: number | undefined;
	maxRetailPrice: number | undefined;
	model: '' | 'FIXED_DEPOSIT' | 'COMMISSION';
	commissionPercent: string;
}

export type NeedErrors = Partial<Record<keyof NeedValues, string>>;

export const fieldsFromServer: Record<string, keyof NeedValues | undefined> = {
	title: 'title',
	category_id: 'categoryId',
	description: 'description',
	est_qty_per_week: 'weeklyQuantity',
	max_retail_price: 'maxRetailPrice',
	preferred_model: 'model',
	offered_commission_bps: 'commissionPercent'
};

export function emptyNeedValues(): NeedValues {
	return {
		title: '',
		categoryId: '',
		description: '',
		weeklyQuantity: undefined,
		maxRetailPrice: undefined,
		model: '',
		commissionPercent: ''
	};
}

export function validateNeed(values: NeedValues): NeedErrors {
	const errors: NeedErrors = {};
	const titleLength = Array.from(values.title.trim()).length;
	if (titleLength < minTitleLength || titleLength > maxTitleLength) {
		errors.title = id.needs.titleLength;
	}
	if (values.categoryId === '') {
		errors.categoryId = id.needs.chooseCategory;
	}
	return errors;
}

function commissionOf(values: NeedValues): number | null {
	const percent = Number(values.commissionPercent.replace(',', '.'));
	if (values.commissionPercent.trim() === '' || !Number.isFinite(percent)) {
		return null;
	}
	return Math.round(percent * basisPointsPerPercent);
}

export function toNeedRequest(values: NeedValues): NeedRequest {
	return {
		title: values.title.trim(),
		category_id: values.categoryId,
		description: values.description,
		est_qty_per_week: values.weeklyQuantity ?? null,
		max_retail_price: values.maxRetailPrice ?? null,
		preferred_model: values.model === '' ? null : values.model,
		offered_commission_bps: commissionOf(values)
	};
}
