import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { reportListing, type ListingReportRequest } from '#lib/api/listing-reports.js';
import { listingReport as id } from '#lib/i18n/listing-report.js';

export type ReportReason = ListingReportRequest['reason'];
export type ReportTarget = Pick<ListingReportRequest, 'target_type' | 'target_id'>;
export type ReportOutcome =
	{ kind: 'sent' } | { kind: 'invalid' } | { kind: 'failed'; message: string };

export class ListingReportForm {
	reason = $state<ReportReason | ''>('');
	note = $state('');
	errors = $state<Record<string, string>>({});
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async send(target: ReportTarget): Promise<ReportOutcome> {
		this.errors = this.#problems();
		if (this.reason === '' || Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			await reportListing(this.#client(), {
				...target,
				reason: this.reason,
				note: this.note.trim()
			});
			this.reason = '';
			this.note = '';
			return { kind: 'sent' };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	#problems(): Record<string, string> {
		const text = id.listingReport;
		if (this.reason === '') {
			return { reason: text.reasonRequired };
		}
		return this.reason === 'OTHER' && this.note.trim() === '' ? { note: text.noteRequired } : {};
	}

	#explain(error: unknown): ReportOutcome {
		if (error instanceof ApiError && error.fields.length > 0) {
			this.errors = Object.fromEntries(error.fields.map((field) => [field.field, field.message]));
			return { kind: 'invalid' };
		}
		return { kind: 'failed', message: error instanceof Error ? error.message : '' };
	}
}
