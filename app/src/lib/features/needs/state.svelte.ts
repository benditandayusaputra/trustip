import { untrack } from 'svelte';
import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import {
	changeNeed,
	getNeed,
	listNeeds,
	postNeed,
	type Need,
	type NeedAction,
	type NeedDetail,
	type NeedStatus
} from '#lib/api/needs.js';
import {
	emptyNeedValues,
	fieldsFromServer,
	toNeedRequest,
	validateNeed,
	type NeedErrors,
	type NeedValues
} from './schema';

export type PostOutcome =
	{ kind: 'posted'; need: Need } | { kind: 'invalid' } | { kind: 'failed'; message: string };

export class NeedComposer {
	values = $state<NeedValues>(emptyNeedValues());
	errors = $state<NeedErrors>({});
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async post(): Promise<PostOutcome> {
		this.errors = validateNeed(this.values);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			return { kind: 'posted', need: await postNeed(this.#client(), toNeedRequest(this.values)) };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	#explain(error: unknown): PostOutcome {
		if (!(error instanceof ApiError) || error.fields.length === 0) {
			return { kind: 'failed', message: error instanceof Error ? error.message : '' };
		}
		const errors: NeedErrors = {};
		for (const problem of error.fields) {
			const name = fieldsFromServer[problem.field];
			if (name) {
				errors[name] = problem.message;
			}
		}
		this.errors = errors;
		return { kind: 'invalid' };
	}
}

export class NeedList {
	items = $state<Need[]>([]);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	#filter: NeedStatus | '' = '';
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(filter: NeedStatus | ''): Promise<void> {
		this.#filter = filter;
		if (untrack(() => this.status) !== 'ready') {
			this.status = 'loading';
		}
		try {
			const page = await listNeeds(this.#client(), filter, null);
			this.items = page.items;
			this.nextCursor = page.hasMore ? page.nextCursor : null;
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async loadMore(): Promise<void> {
		if (this.nextCursor === null) {
			return;
		}
		this.busy = true;
		try {
			const page = await listNeeds(this.#client(), this.#filter, this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}
}

export class NeedRecord {
	detail = $state<NeedDetail>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(needId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.detail = await getNeed(this.#client(), needId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async change(action: NeedAction): Promise<string | undefined> {
		const current = this.detail;
		if (!current) {
			return undefined;
		}
		this.busy = true;
		try {
			const changed = await changeNeed(this.#client(), current.id, action);
			this.detail = { ...current, ...changed };
			return undefined;
		} catch (error) {
			return error instanceof Error ? error.message : '';
		} finally {
			this.busy = false;
		}
	}
}
