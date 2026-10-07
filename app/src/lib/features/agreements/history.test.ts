import { describe, expect, it } from 'vitest';
import { describeItemChange, describeTermChange } from './history';

const commission = {
	pricing_model: 'COMMISSION' as const,
	retail_price: 15000,
	commission_bps: 2000,
	deposit_price: 0
};

describe('agreement history', () => {
	it('FR-AGR-04 describes a changed general term with readable values', () => {
		expect(
			describeTermChange({ field: 'settlement_schedule', before: 'WEEKLY', after: 'MONTHLY' })
		).toBe('Jadwal setor: Mingguan, Senin sampai Minggu menjadi Bulanan, per bulan kalender');
	});

	it('FR-AGR-04 describes added, updated, and stopped products', () => {
		expect(
			describeItemChange({
				product_id: 'p',
				product_name: 'Keripik',
				kind: 'ADDED',
				before: null,
				after: commission
			})
		).toBe('Ditambah: Keripik (Komisi 20%, Rp15.000)');
		expect(
			describeItemChange({
				product_id: 'p',
				product_name: 'Keripik',
				kind: 'UPDATED',
				before: commission,
				after: { ...commission, commission_bps: 1500 }
			})
		).toBe('Diubah: Keripik (Komisi 20%, Rp15.000 → Komisi 15%, Rp15.000)');
		expect(
			describeItemChange({
				product_id: 'p',
				product_name: 'Roti',
				kind: 'STOPPED',
				before: commission,
				after: commission
			})
		).toBe('Dihentikan: Roti');
	});
});
