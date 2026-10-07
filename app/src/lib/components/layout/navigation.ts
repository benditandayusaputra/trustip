import type { Component } from 'svelte';
import Calculator from '@lucide/svelte/icons/calculator';
import ChartColumn from '@lucide/svelte/icons/chart-column';
import Bell from '@lucide/svelte/icons/bell';
import CloudUpload from '@lucide/svelte/icons/cloud-upload';
import FileSignature from '@lucide/svelte/icons/file-signature';
import Handshake from '@lucide/svelte/icons/handshake';
import House from '@lucide/svelte/icons/house';
import Megaphone from '@lucide/svelte/icons/megaphone';
import Package from '@lucide/svelte/icons/package';
import PackageOpen from '@lucide/svelte/icons/package-open';
import Search from '@lucide/svelte/icons/search';
import Settings from '@lucide/svelte/icons/settings';
import Warehouse from '@lucide/svelte/icons/warehouse';
import Wallet from '@lucide/svelte/icons/wallet';
import { id } from '#lib/i18n/id.js';

export type OrganizationType = 'CONSIGNOR' | 'STORE';

export interface NavItem {
	href: string;
	label: string;
	icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
}

interface Navigation {
	primary: NavItem[];
	more: NavItem[];
}

const shared: NavItem[] = [
	{ href: '/app/ajuan', label: id.nav.proposals, icon: FileSignature },
	{ href: '/app/stok', label: id.nav.stock, icon: Warehouse },
	{ href: '/app/laporan', label: id.nav.reports, icon: ChartColumn },
	{ href: '/app/pemberitahuan', label: id.nav.notifications, icon: Bell },
	{ href: '/app/antrean', label: id.nav.outbox, icon: CloudUpload },
	{ href: '/app/pengaturan', label: id.nav.settings, icon: Settings }
];

const byType: Record<OrganizationType, Navigation> = {
	CONSIGNOR: {
		primary: [
			{ href: '/app/penitip', label: id.nav.home, icon: House },
			{ href: '/cari/toko', label: id.nav.search, icon: Search },
			{ href: '/app/mitra', label: id.nav.partners, icon: Handshake },
			{ href: '/app/setoran', label: id.nav.settlements, icon: Wallet }
		],
		more: [{ href: '/app/produk', label: id.nav.products, icon: PackageOpen }, ...shared]
	},
	STORE: {
		primary: [
			{ href: '/app/toko', label: id.nav.home, icon: House },
			{ href: '/app/hitung', label: id.nav.count, icon: Calculator },
			{ href: '/app/titipan', label: id.nav.drops, icon: Package },
			{ href: '/cari/penitip', label: id.nav.search, icon: Search }
		],
		more: [
			{ href: '/app/mitra', label: id.nav.partners, icon: Handshake },
			{ href: '/app/setoran', label: id.nav.settlements, icon: Wallet },
			{ href: '/app/kebutuhan', label: id.nav.needs, icon: Megaphone },
			...shared
		]
	}
};

export function navigationFor(type: OrganizationType): Navigation {
	return byType[type];
}

export function isCurrent(href: string, currentPath: string): boolean {
	return currentPath === href || currentPath.startsWith(`${href}/`);
}
