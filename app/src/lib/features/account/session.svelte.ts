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
const workspaceKey = 'trustip-workspace';
const workspaceTypes: readonly string[] = ['CONSIGNOR', 'STORE'] satisfies OrganizationType[];

function readWorkspace(): OrganizationType | undefined {
	try {
		const stored = localStorage.getItem(workspaceKey) ?? '';
		return workspaceTypes.includes(stored) ? (stored as OrganizationType) : undefined;
	} catch {
		return undefined;
	}
}

function writeWorkspace(type: OrganizationType | undefined): void {
	try {
		if (type) {
			localStorage.setItem(workspaceKey, type);
		} else {
			localStorage.removeItem(workspaceKey);
		}
	} catch {
		return;
	}
}
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
	expectedWorkspace = $state<OrganizationType | undefined>();
	readonly #client: () => ApiClient;
	readonly #forgetDeviceData: () => Promise<void>;

	constructor(client: () => ApiClient, forgetDeviceData = () => deviceReadCache().forget()) {
		this.#client = client;
		this.#forgetDeviceData = forgetDeviceData;
	}

	async load(): Promise<void> {
		if (this.status !== 'ready') {
			this.status = 'loading';
		}
		try {
			this.account = await getAccount(this.#client());
			this.status = 'ready';
			writeWorkspace(this.account.membership?.organization_type);
		} catch (error) {
			this.#fail(error);
		}
	}

	recallWorkspace(): void {
		this.expectedWorkspace = readWorkspace();
	}

	async endThisDevice(): Promise<void> {
		await logOut(this.#client());
		writeWorkspace(undefined);
		await this.#forgetDeviceData();
		this.account = undefined;
		this.status = 'signedOut';
	}

	async endEveryDevice(): Promise<void> {
		await logOutEverywhere(this.#client());
		writeWorkspace(undefined);
		await this.#forgetDeviceData();
		this.account = undefined;
		this.status = 'signedOut';
	}

	#fail(error: unknown): void {
		if (error instanceof ApiError && error.isSessionProblem()) {
			writeWorkspace(undefined);
			this.account = undefined;
			this.status = 'signedOut';
			return;
		}
		this.failure = error instanceof Error ? error.message : '';
		this.status = 'failed';
	}
}

export const [getAccountSession, setAccountSession] = createContext<AccountSession>();
