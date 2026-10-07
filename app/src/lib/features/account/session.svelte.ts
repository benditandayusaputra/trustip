import { createContext } from 'svelte';
import type { ApiClient } from '#lib/api/client.js';
import { getAccount } from '#lib/api/account.js';
import { ApiError } from '#lib/api/errors.js';
import { logOut, logOutEverywhere } from '#lib/api/sessions.js';
import type { OrganizationType } from '#lib/api/organizations.js';
import type { Account } from '#lib/api/users.js';

export type AccountStatus = 'loading' | 'ready' | 'failed' | 'signedOut';
export type Placement = 'member' | 'newcomer';

const dashboards: Record<OrganizationType, string> = {
	CONSIGNOR: '/app/penitip',
	STORE: '/app/toko'
};

export function dashboardPath(type: OrganizationType): string {
	return dashboards[type];
}

export function isPlacedFor(account: Account, placement: Placement): boolean {
	return (account.membership !== null) === (placement === 'member');
}

export class AccountSession {
	account = $state<Account | undefined>();
	status = $state<AccountStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
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
		this.status = 'signedOut';
	}

	async endEveryDevice(): Promise<void> {
		await logOutEverywhere(this.#client());
		this.status = 'signedOut';
	}

	#fail(error: unknown): void {
		if (error instanceof ApiError && error.isSessionProblem()) {
			this.status = 'signedOut';
			return;
		}
		this.failure = error instanceof Error ? error.message : '';
		this.status = 'failed';
	}
}

export const [getAccountSession, setAccountSession] = createContext<AccountSession>();
