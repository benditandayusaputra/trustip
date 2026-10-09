import { error } from '@sveltejs/kit';
import { ApiError } from '#lib/api/errors.js';
import { id } from '#lib/i18n/id.js';

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

export async function findPublic<T>(read: () => Promise<T>): Promise<T | null> {
	try {
		return await read();
	} catch (failure) {
		if (isNotFound(failure)) {
			return null;
		}
		throw failure;
	}
}
