type DemoPersona = 'bendi' | 'zara' | 'jessica';

export interface DemoAccount {
	persona: DemoPersona;
	email: string;
	password: string;
}

const demoPassword = 'Coba-Demo-2026';

export const demoAccounts: readonly DemoAccount[] = [
	{ persona: 'bendi', email: 'penitip@demo.trustip.id', password: demoPassword },
	{ persona: 'zara', email: 'toko@demo.trustip.id', password: demoPassword },
	{ persona: 'jessica', email: 'minimarket@demo.trustip.id', password: demoPassword }
];
