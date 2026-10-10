import type { Product, ProductRequest } from '#lib/api/products.js';
import { common as commonText } from '#lib/i18n/common.js';
import { organization as organizationText } from '#lib/i18n/organization.js';
import { products as productsText } from '#lib/i18n/products.js';

const id = { ...commonText, ...organizationText, ...productsText };

const minNameLength = 2;
const maxNameLength = 80;
const maxUnitLength = 20;
const defaultShelfLifeDays = 3;

export interface ProductValues {
	name: string;
	sku: string;
	categoryId: string;
	unit: string;
	description: string;
	composition: string;
	netContent: string;
	retailPrice: number | undefined;
	depositPrice: number | undefined;
	hasShelfLife: boolean;
	shelfLifeDays: number;
}

export type ProductField = keyof ProductValues | 'photo';
export type ProductErrors = Partial<Record<ProductField, string>>;

export const fieldsFromServer: Record<string, ProductField | undefined> = {
	name: 'name',
	sku: 'sku',
	category_id: 'categoryId',
	unit: 'unit',
	description: 'description',
	composition: 'composition',
	net_content: 'netContent',
	photo_file_id: 'photo',
	suggested_retail_price: 'retailPrice',
	suggested_deposit_price: 'depositPrice',
	shelf_life_days: 'shelfLifeDays'
};

export function emptyProductValues(): ProductValues {
	return {
		name: '',
		sku: '',
		categoryId: '',
		unit: 'pcs',
		description: '',
		composition: '',
		netContent: '',
		retailPrice: undefined,
		depositPrice: undefined,
		hasShelfLife: true,
		shelfLifeDays: defaultShelfLifeDays
	};
}

export function toProductValues(product: Product): ProductValues {
	return {
		name: product.name,
		sku: product.sku,
		categoryId: product.category_id,
		unit: product.unit,
		description: product.description,
		composition: product.composition,
		netContent: product.net_content,
		retailPrice: product.suggested_retail_price,
		depositPrice: product.suggested_deposit_price,
		hasShelfLife: product.shelf_life_days !== null,
		shelfLifeDays: product.shelf_life_days ?? defaultShelfLifeDays
	};
}

export function toProductRequest(values: ProductValues, photoId: string | null): ProductRequest {
	return {
		sku: values.sku.trim(),
		name: values.name.trim(),
		category_id: values.categoryId,
		unit: values.unit.trim(),
		description: values.description,
		composition: values.composition,
		net_content: values.netContent,
		photo_file_id: photoId,
		suggested_retail_price: values.retailPrice ?? 0,
		suggested_deposit_price: values.depositPrice ?? 0,
		shelf_life_days: values.hasShelfLife ? values.shelfLifeDays : null
	};
}

function lengthProblem(value: string, min: number, max: number): string | undefined {
	const length = Array.from(value.trim()).length;
	if (length === 0) {
		return id.auth.validation.required;
	}
	if (length < min) {
		return id.auth.validation.minLength(min);
	}
	return length > max ? id.products.maxLength(max) : undefined;
}

export function validateProduct(values: ProductValues): ProductErrors {
	const problems: [ProductField, string | undefined][] = [
		['name', lengthProblem(values.name, minNameLength, maxNameLength)],
		['unit', lengthProblem(values.unit, 1, maxUnitLength)],
		['categoryId', values.categoryId === '' ? id.onboarding.validation.choose : undefined],
		['depositPrice', priceProblem(values)]
	];
	return Object.fromEntries(problems.filter(([, message]) => message !== undefined));
}

function priceProblem(values: ProductValues): string | undefined {
	const isOrdered = (values.depositPrice ?? 0) <= (values.retailPrice ?? 0);
	return isOrdered ? undefined : id.products.priceOrder;
}
