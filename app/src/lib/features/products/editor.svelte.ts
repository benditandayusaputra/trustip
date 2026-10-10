import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { uploadImage } from '#lib/api/files.js';
import { listCategories, type Category } from '#lib/api/organizations.js';
import {
	createProduct,
	updateProduct,
	type Product,
	type ProductRequest
} from '#lib/api/products.js';
import { organization as id } from '#lib/i18n/organization.js';
import {
	emptyProductValues,
	fieldsFromServer,
	toProductRequest,
	toProductValues,
	validateProduct,
	type ProductErrors,
	type ProductValues
} from './schema';

export type ProductOutcome =
	{ kind: 'saved'; product: Product } | { kind: 'invalid' } | { kind: 'failed'; message: string };

export class ProductEditor {
	values = $state<ProductValues>(emptyProductValues());
	photo = $state<Product['photo']>(null);
	categories = $state<Category[]>([]);
	errors = $state<ProductErrors>({});
	busy = $state(false);
	isUploading = $state(false);
	readonly #client: () => ApiClient;
	readonly #productId: string | null;

	constructor(client: () => ApiClient, product: Product | null) {
		this.#client = client;
		this.#productId = product?.id ?? null;
		if (product) {
			this.values = toProductValues(product);
			this.photo = product.photo;
		}
	}

	async loadCategories(): Promise<void> {
		this.categories = await listCategories(this.#client()).catch(() => []);
	}

	async choosePhoto(file: File): Promise<void> {
		this.isUploading = true;
		this.errors = { ...this.errors, photo: '' };
		try {
			const stored = await uploadImage(this.#client(), file, 'PRODUCT_PHOTO');
			this.photo = { id: stored.id, url: stored.url };
		} catch (error) {
			this.errors.photo = error instanceof ApiError ? error.message : id.settings.imageInvalid;
		} finally {
			this.isUploading = false;
		}
	}

	async save(): Promise<ProductOutcome> {
		this.errors = validateProduct(this.values);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			const product = await this.#send(toProductRequest(this.values, this.photo?.id ?? null));
			return { kind: 'saved', product };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	#send(request: ProductRequest): Promise<Product> {
		if (this.#productId === null) {
			return createProduct(this.#client(), request);
		}
		return updateProduct(this.#client(), this.#productId, request);
	}

	#explain(error: unknown): ProductOutcome {
		if (!(error instanceof ApiError) || error.fields.length === 0) {
			return { kind: 'failed', message: error instanceof Error ? error.message : '' };
		}
		const errors: ProductErrors = {};
		error.fields.forEach((problem) => {
			const field = fieldsFromServer[problem.field];
			if (field) {
				errors[field] = problem.message;
			}
		});
		this.errors = errors;
		return { kind: 'invalid' };
	}
}
