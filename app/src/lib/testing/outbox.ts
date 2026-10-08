import { outboxScheduler } from '#lib/offline/device.js';
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

export async function emptyOutbox(): Promise<void> {
	outboxScheduler().stop();
	for (const entry of await outboxStore.listFor(storeCashier.id)) {
		await outboxStore.remove(entry.id);
	}
}
