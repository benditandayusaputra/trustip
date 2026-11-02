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
import UserRound from '@lucide/svelte/icons/user-round';
import Warehouse from '@lucide/svelte/icons/warehouse';
import Wallet from '@lucide/svelte/icons/wallet';
import { common as id } from '#lib/i18n/common.js';

export type OrganizationType = 'CONSIGNOR' | 'STORE';

export interface NavItem {
	href: string;
	label: string;
	icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
}

export interface Navigation {
	workspace: NavItem[];
	manage: NavItem[];
}

const reports: NavItem = { href: '/app/laporan', label: id.nav.reports, icon: ChartColumn };
const partners: NavItem = { href: '/app/mitra', label: id.nav.partners, icon: Handshake };
const proposals: NavItem = { href: '/app/ajuan', label: id.nav.proposals, icon: FileSignature };
const stock: NavItem = { href: '/app/stok', label: id.nav.stock, icon: Warehouse };
const settings: NavItem = { href: '/app/pengaturan', label: id.nav.settings, icon: Settings };

const byType: Record<OrganizationType, Navigation> = {
	CONSIGNOR: {
		workspace: [
			{ href: '/app/penitip', label: id.nav.home, icon: House },
			{ href: '/cari/toko', label: id.nav.search, icon: Search },
			partners,
			reports
		],
		manage: [
			{ href: '/app/produk', label: id.nav.products, icon: PackageOpen },
			proposals,
			stock,
			settings
		]
	},
	STORE: {
		workspace: [
			{ href: '/app/toko', label: id.nav.home, icon: House },
			{ href: '/app/hitung', label: id.nav.count, icon: Calculator },
			{ href: '/app/titipan', label: id.nav.drops, icon: Package },
			{ href: '/cari/penitip', label: id.nav.search, icon: Search },
			reports
		],
		manage: [
			partners,
			{ href: '/app/kebutuhan', label: id.nav.needs, icon: Megaphone },
			proposals,
			stock,
			{ href: '/app/antrean', label: id.nav.outbox, icon: CloudUpload },
			settings
		]
	}
};

export const sectionsOutsideNavigation: NavItem[] = [
	{ href: '/app/hitung', label: id.nav.count, icon: Calculator },
	{ href: '/app/titipan', label: id.nav.drops, icon: Package },
	{ href: '/app/setoran', label: id.nav.settlements, icon: Wallet },
	{ href: '/app/pemberitahuan', label: id.nav.notifications, icon: Bell },
	{ href: '/app/profil', label: id.nav.profile, icon: UserRound },
	{ href: '/app/antrean', label: id.nav.outbox, icon: CloudUpload },
	{ href: '/kebutuhan', label: id.nav.needsBoard, icon: Megaphone },
	{ href: '/cari/toko', label: id.nav.findStores, icon: Search },
	{ href: '/cari/penitip', label: id.nav.findConsignors, icon: Search }
];

const menuOfPagesOutside: Record<OrganizationType, Record<string, string>> = {
	CONSIGNOR: {
		'/app/hitung': '/app/mitra',
		'/app/titipan': '/app/mitra',
		'/app/setoran': '/app/laporan',
		'/cari/penitip': '/cari/toko',
		'/kebutuhan': '/cari/toko',
		'/t': '/cari/toko',
		'/u': '/cari/toko',
		'/p': '/cari/toko'
	},
	STORE: {
		'/app/setoran': '/app/laporan',
		'/cari/toko': '/cari/penitip',
		'/kebutuhan': '/app/kebutuhan',
		'/t': '/cari/penitip',
		'/u': '/cari/penitip',
		'/p': '/cari/penitip'
	}
};

export function navigationFor(type: OrganizationType): Navigation {
	return byType[type];
}

export function menuOfPage(type: OrganizationType, path: string): string | undefined {
	return Object.entries(menuOfPagesOutside[type]).find(([prefix]) => isCurrent(prefix, path))?.[1];
}

export function isCurrent(href: string, currentPath: string): boolean {
	const path = href.split('?')[0] ?? href;
	return currentPath === path || currentPath.startsWith(`${path}/`);
}

const focusPaths = ['/app/pemberitahuan', '/app/profil'];

export function isFocusPage(currentPath: string): boolean {
	return focusPaths.some((path) => isCurrent(path, currentPath));
}

const bottomTabs = 4;

export function bottomNavigationOf(navigation: Navigation): {
	primary: NavItem[];
	more: NavItem[];
} {
	const all = [...navigation.workspace, ...navigation.manage];
	return { primary: all.slice(0, bottomTabs), more: all.slice(bottomTabs) };
}
