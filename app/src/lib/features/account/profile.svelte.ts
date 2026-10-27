import type { ApiClient } from '#lib/api/client.js';
import {
	changePassword,
	updateAccount,
	type AccountProfile,
	type AccountUpdateRequest
} from '#lib/api/account.js';
import { ApiError } from '#lib/api/errors.js';
import { uploadImage } from '#lib/api/files.js';
import { localWhatsApp } from '#lib/domain/whatsapp.js';

export type AccountField = 'fullName' | 'whatsapp' | 'avatar' | 'currentPassword' | 'newPassword';
export type AccountOutcome =
	{ kind: 'saved' } | { kind: 'invalid' } | { kind: 'failed'; message: string };

const fieldsFromServer: Record<string, AccountField | undefined> = {
	full_name: 'fullName',
	whatsapp: 'whatsapp',
	avatar_file_id: 'avatar',
	file: 'avatar',
	current_password: 'currentPassword',
	new_password: 'newPassword'
};

export class AccountEditor {
	fullName = $state('');
	whatsapp = $state('');
	errors = $state<Partial<Record<AccountField, string>>>({});
	busy = $state(false);
	readonly #client: () => ApiClient;
	readonly #onsaved: (account: AccountProfile) => void;
	#account: AccountProfile;

	constructor(
		client: () => ApiClient,
		account: AccountProfile,
		onsaved: (account: AccountProfile) => void
	) {
		this.#client = client;
		this.#account = account;
		this.#onsaved = onsaved;
		this.reset();
	}

	reset(): void {
		this.fullName = this.#account.full_name;
		this.whatsapp = localWhatsApp(this.#account.whatsapp);
		this.errors = {};
	}

	save(): Promise<AccountOutcome> {
		return this.#write({ full_name: this.fullName, whatsapp: this.whatsapp });
	}

	async chooseAvatar(file: File): Promise<AccountOutcome> {
		this.busy = true;
		this.errors = {};
		try {
			const stored = await uploadImage(this.#client(), file, 'AVATAR');
			return await this.#write({ avatar_file_id: stored.id });
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	removeAvatar(): Promise<AccountOutcome> {
		return this.#write({ avatar_file_id: null });
	}

	async changePassword(current: string, next: string): Promise<AccountOutcome> {
		this.busy = true;
		this.errors = {};
		try {
			await changePassword(this.#client(), { current_password: current, new_password: next });
			return { kind: 'saved' };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	async #write(change: Partial<AccountUpdateRequest>): Promise<AccountOutcome> {
		this.busy = true;
		this.errors = {};
		const account = this.#account;
		try {
			const saved = await updateAccount(this.#client(), {
				full_name: account.full_name,
				whatsapp: account.whatsapp,
				avatar_file_id: account.avatar?.id ?? null,
				...change
			});
			this.#account = { ...saved, membership: account.membership };
			this.reset();
			this.#onsaved(this.#account);
			return { kind: 'saved' };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	#explain(error: unknown): AccountOutcome {
		if (!(error instanceof ApiError) || error.fields.length === 0) {
			return { kind: 'failed', message: error instanceof Error ? error.message : '' };
		}
		const errors: Partial<Record<AccountField, string>> = {};
		for (const problem of error.fields) {
			const field = fieldsFromServer[problem.field];
			if (field) {
				errors[field] = problem.message;
			}
		}
		this.errors = errors;
		return { kind: 'invalid' };
	}
}
