import { createContext } from 'svelte';
import type { ApiClient } from '#lib/api/client.js';
import { getAccount, type AccountProfile } from '#lib/api/account.js';
import { ApiError } from '#lib/api/errors.js';
import { logOut, logOutEverywhere } from '#lib/api/sessions.js';
import type { OrganizationType } from '#lib/api/organizations.js';
import type { Account } from '#lib/api/users.js';
import { dashboardDestination } from '#lib/domain/dashboard-access.js';
import { deviceReadCache } from '#lib/offline/read-cache.js';

export type AccountStatus = 'loading' | 'ready' | 'failed' | 'signedOut';
export type Placement = 'member' | 'newcomer' | 'anyone';

export function dashboardPath(type: OrganizationType): string {
	return dashboardDestination('/app', type);
}

export function isPlacedFor(account: Account, placement: Placement): boolean {
	return (account.membership !== null) === (placement !== 'newcomer');
}

export class AccountSession {
	account = $state<AccountProfile | undefined>();
	status = $state<AccountStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;
	readonly #forgetDeviceData: () => Promise<void>;

	constructor(client: () => ApiClient, forgetDeviceData = () => deviceReadCache().forget()) {
		this.#client = client;
		this.#forgetDeviceData = forgetDeviceData;
	}

	async load(): Promise<void> {
		this.status = 'loading';
		try {
			this.account = await getAccount(this.#client());
			this.status = 'ready';
		} catch (error) {
			this.#fail(error);
		}
	}

	async endThisDevice(): Promise<void> {
		await logOut(this.#client());
		await this.#forgetDeviceData();
		this.account = undefined;
		this.status = 'signedOut';
	}

	async endEveryDevice(): Promise<void> {
		await logOutEverywhere(this.#client());
		await this.#forgetDeviceData();
		this.account = undefined;
		this.status = 'signedOut';
	}

	#fail(error: unknown): void {
		if (error instanceof ApiError && error.isSessionProblem()) {
			this.account = undefined;
			this.status = 'signedOut';
			return;
		}
		this.failure = error instanceof Error ? error.message : '';
		this.status = 'failed';
	}
}

export const [getAccountSession, setAccountSession] = createContext<AccountSession>();
