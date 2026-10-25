import { error } from '@sveltejs/kit';
import { ApiError } from '#lib/api/errors.js';
import { common as id } from '#lib/i18n/common.js';

const notFoundStatus = 404;

function isNotFound(failure: unknown): boolean {
	return failure instanceof ApiError && failure.status === notFoundStatus;
}

export async function readPublic<T>(read: () => Promise<T>): Promise<T> {
	try {
		return await read();
	} catch (failure) {
		if (isNotFound(failure)) {
			error(notFoundStatus, id.errorPage.notFoundBody);
		}
		throw failure;
	}
}
