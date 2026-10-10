import type { ProposalRevisionRecord } from '#lib/api/proposals.js';
import { proposals as id } from '#lib/i18n/proposals.js';

const text = id.proposals;

function namesWith(record: ProposalRevisionRecord, change: string): string {
	return record.items
		.filter((item) => item.change === change)
		.map((item) => item.product_name)
		.join(', ');
}

export function revisionSummary(record: ProposalRevisionRecord): string[] {
	if (record.revision_no === 1) {
		return [text.firstRevision];
	}
	const terms = record.term_changes.map((field) => text.termNames[field]).join(', ');
	const lines: [string, (names: string) => string][] = [
		[terms, text.termsChanged],
		[namesWith(record, 'CHANGED'), text.itemsChanged],
		[namesWith(record, 'ADDED'), text.itemsAdded],
		[namesWith(record, 'REMOVED'), text.itemsRemoved]
	];
	return lines.filter(([names]) => names !== '').map(([names, describe]) => describe(names));
}
