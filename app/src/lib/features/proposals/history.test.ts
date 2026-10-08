import { describe, expect, it } from 'vitest';
import type { ProposalRevisionRecord } from '#lib/api/proposals.js';
import { revisionSummary } from './history';

function item(name: string, change: 'ADDED' | 'CHANGED' | 'KEPT' | 'REMOVED') {
	return {
		product_id: name,
		product_name: name,
		sku: 'S',
		unit: 'pcs',
		action: 'ADD' as const,
		pricing_model: 'FIXED_DEPOSIT' as const,
		retail_price: 2000,
		commission_bps: 0,
		deposit_price: 1700,
		per_unit: { gross: 2000, store_share: 300, consignor_share: 1700 },
		change
	};
}

function record(overrides: Partial<ProposalRevisionRecord>): ProposalRevisionRecord {
	return {
		revision_no: 2,
		author_org_id: 'o',
		author_name: 'Warung Bu Sri',
		terms: {
			settlement_schedule: 'WEEKLY',
			payment_terms_days: 14,
			count_rhythm: 'DAILY',
			notes: ''
		},
		term_changes: [],
		message: '',
		items: [],
		created_at: '2026-10-21T08:00:00Z',
		...overrides
	};
}

describe('revision summary', () => {
	it('FR-PRP-04 names what changed against the previous revision', () => {
		const summary = revisionSummary(
			record({
				term_changes: ['payment_terms_days'],
				items: [item('Roti Manis', 'CHANGED'), item('Bolu', 'ADDED'), item('Keripik', 'REMOVED')]
			})
		);

		expect(summary).toEqual([
			'Syarat berubah: tempo bayar',
			'Produk berubah: Roti Manis',
			'Produk baru: Bolu',
			'Produk dihapus: Keripik'
		]);
	});

	it('labels the first revision as the opening proposal', () => {
		expect(revisionSummary(record({ revision_no: 1, items: [item('Roti', 'ADDED')] }))).toEqual([
			'Ajuan awal'
		]);
	});
});
