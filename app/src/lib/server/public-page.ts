import { error } from '@sveltejs/kit';
import { ApiError } from '#lib/api/errors.js';
import { id } from '#lib/i18n/id.js';

const notFoundStatus = 404;

export async function readPublic<T>(read: () => Promise<T>): Promise<T> {
	try {
		return await read();
	} catch (failure) {
		if (failure instanceof ApiError && failure.status === notFoundStatus) {
			error(notFoundStatus, id.errorPage.notFoundBody);
		}
		throw failure;
	}
}
