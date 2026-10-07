import type { ApiClient } from '#lib/api/client.js';
import { setListing, type Organization } from '#lib/api/organizations.js';
import type { SectionOutcome } from './section';

export class ListingEditor {
	isListed = $state(false);
	busy = $state(false);
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient, organization: Organization) {
		this.#client = client;
		this.isListed = organization.listing?.is_listed ?? false;
	}

	async change(organizationId: string, isListed: boolean): Promise<SectionOutcome> {
		this.busy = true;
		this.failure = '';
		try {
			const saved = await setListing(this.#client(), organizationId, isListed);
			this.isListed = saved.listing?.is_listed ?? isListed;
			return { kind: 'saved', organization: saved };
		} catch (error) {
			this.isListed = !isListed;
			this.failure = error instanceof Error ? error.message : '';
			return { kind: 'invalid' };
		} finally {
			this.busy = false;
		}
	}
}
