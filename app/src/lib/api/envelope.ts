import type { components } from './schema';
import { ApiError, type FieldError } from './errors';
import { id } from '#lib/i18n/id.js';

type ErrorResponse = components['schemas']['ErrorResponse'];
type ListMeta = components['schemas']['ListMeta'];

export interface SuccessEnvelope<T> {
	code: 'OK';
	message: string;
	data: T;
	meta: { request_id: string } | ListMeta;
}

export interface Page<T> {
	items: T[];
	nextCursor: string | null;
	hasMore: boolean;
	search?: unknown;
}

export async function readEnvelope<T>(response: Response): Promise<SuccessEnvelope<T>> {
	const body: unknown = await response.json().catch(() => undefined);
	if (response.ok && isEnvelope(body)) {
		return body as SuccessEnvelope<T>;
	}
	throw toApiError(response, body);
}

export function toPage<T>(envelope: SuccessEnvelope<T[]>): Page<T> {
	const meta = envelope.meta as Partial<ListMeta> & { search?: unknown };
	return {
		items: envelope.data,
		nextCursor: meta.next_cursor ?? null,
		hasMore: meta.has_more ?? false,
		...(meta.search === undefined ? {} : { search: meta.search })
	};
}

function isEnvelope(body: unknown): body is { code: string; data: unknown } {
	return typeof body === 'object' && body !== null && 'code' in body;
}

type FailureBody = ErrorResponse & { errors?: FieldError[] };

function toApiError(response: Response, body: unknown): ApiError {
	const failure = isEnvelope(body) ? (body as unknown as FailureBody) : undefined;
	return new ApiError({
		status: response.status,
		code: failure?.code ?? 'INTERNAL',
		message: failure?.message ?? id.apiErrors.unreadable,
		requestId: requestIdOf(response, failure),
		fields: failure?.errors ?? [],
		retryAfterSeconds: retryAfterOf(response)
	});
}

function requestIdOf(response: Response, failure: FailureBody | undefined): string {
	return failure?.meta.request_id ?? response.headers.get('x-request-id') ?? '';
}

function retryAfterOf(response: Response): number | undefined {
	const seconds = Number(response.headers.get('retry-after'));
	return Number.isFinite(seconds) && seconds > 0 ? seconds : undefined;
}
