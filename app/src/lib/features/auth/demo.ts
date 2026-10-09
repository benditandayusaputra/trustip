export interface DemoAccount {
	role: 'consignor' | 'store';
	email: string;
	password: string;
}

export const demoAccounts: readonly DemoAccount[] = [
	{ role: 'consignor', email: 'penitip@demo.trustip.id', password: 'Coba-Demo-2026' },
	{ role: 'store', email: 'toko@demo.trustip.id', password: 'Coba-Demo-2026' }
];
