import { outboxScheduler } from '#lib/offline/device.js';
import type { OutboxEntry } from '#lib/offline/outbox.js';
import { outboxStore } from '#lib/offline/store.js';

export const storeCashier = {
	id: '0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d01',
	full_name: 'Sri Lestari',
	email: 'sri@contoh.id',
	membership: {
		organization_id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d',
		organization_type: 'STORE',
		role: 'OWNER'
	}
};

export function aQueuedDrop(changes: Partial<OutboxEntry> = {}): OutboxEntry {
	return {
		id: '0192f0c4-0000-7000-8000-00000000e001',
		revision: 0,
		idempotencyKey: '0192f0c4-0000-7000-8000-00000000e001',
		kind: 'DROP',
		partnershipId: '0192f0c4-0000-7000-8000-000000000501',
		partnerName: 'Roti Pak Anton',
		request: {
			partnership_id: '0192f0c4-0000-7000-8000-000000000501',
			agreement_version: 1,
			occurred_at: null,
			items: []
		},
		lines: [],
		recordedAt: new Date().toISOString(),
		savedOffline: true,
		userId: storeCashier.id,
		organizationId: storeCashier.membership.organization_id,
		status: 'PENDING',
		attempts: 0,
		problem: null,
		document: null,
		...changes
	} as OutboxEntry;
}

export async function emptyOutbox(userId = storeCashier.id): Promise<void> {
	outboxScheduler().stop();
	for (const entry of await outboxStore.listFor(userId)) {
		await outboxStore.remove(entry.id);
	}
}
