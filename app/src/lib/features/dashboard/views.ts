import type { Dashboard } from '#lib/api/insights.js';
import type { SalesNarrowing, SalesReport } from '#lib/api/reports.js';
import { rangeOf, type DateRange } from '#lib/domain/ranges.js';
import { businessDateOf, formatRupiah } from '#lib/i18n/format.js';
import { dashboard as id } from '#lib/i18n/dashboard.js';
import type { RankedRow } from './ranking';
import type { TaskLink } from './components/TaskList.svelte';

const maxRankedRows = 5;
const text = id.dashboard;

interface TaskSpec {
	count: (board: Dashboard) => number;
	label: (count: number) => string;
	href: string;
}

export const consignorTasks: TaskSpec[] = [
	{
		count: (board) => board.tasks.drops_to_confirm,
		label: text.tasks.dropsToConfirm,
		href: '/app/titipan'
	},
	{
		count: (board) => board.tasks.counts_to_confirm,
		label: text.tasks.countsToConfirm,
		href: '/app/hitung'
	},
	{
		count: (board) => board.tasks.payments_to_review,
		label: text.tasks.paymentsToReview,
		href: '/app/setoran'
	},
	{
		count: (board) => board.tasks.proposals_awaiting,
		label: text.tasks.proposalsAwaiting,
		href: '/app/ajuan'
	}
];

export const storeTasks: TaskSpec[] = [
	{
		count: (board) => board.figures.unpaid.count,
		label: text.tasks.unpaidSettlements,
		href: '/app/setoran'
	},
	{
		count: (board) => board.tasks.payments_rejected,
		label: text.tasks.paymentsRejected,
		href: '/app/setoran'
	},
	{
		count: (board) => board.tasks.proposals_awaiting,
		label: text.tasks.proposalsAwaiting,
		href: '/app/ajuan'
	}
];

export function taskLinksOf(board: Dashboard, specs: TaskSpec[]): TaskLink[] {
	return specs
		.map((spec) => ({ count: spec.count(board), label: spec.label, href: spec.href }))
		.filter((task) => task.count > 0)
		.map((task) => ({ label: task.label(task.count), href: task.href }));
}

export function rankedByGross(report: SalesReport | undefined): RankedRow[] {
	return (report?.rows ?? [])
		.filter((row) => row.sold_qty > 0)
		.slice(0, maxRankedRows)
		.map((row) => ({
			key: row.key,
			label: row.label,
			value: formatRupiah(row.split.gross),
			product: row.product
		}));
}

export interface ChartChoice {
	value: string;
	label: string;
}

const partnerPrefix = 'partner:';
const productPrefix = 'product:';

export function chartChoicesOf(
	byPartner: SalesReport | undefined,
	byProduct: SalesReport | undefined
): ChartChoice[] {
	const partners = (byPartner?.rows ?? []).map((row) => ({
		value: partnerPrefix + row.key,
		label: row.label
	}));
	const products = (byProduct?.rows ?? []).map((row) => ({
		value: productPrefix + row.key,
		label: row.label
	}));
	return [{ value: '', label: text.allSales }, ...partners, ...products];
}

export function narrowingOf(choice: string): SalesNarrowing {
	if (choice.startsWith(partnerPrefix)) {
		return { partnerId: choice.slice(partnerPrefix.length) };
	}
	if (choice.startsWith(productPrefix)) {
		return { productId: choice.slice(productPrefix.length) };
	}
	return {};
}

export function lastThirtyDays(now: Date): DateRange {
	const today = businessDateOf(now);
	return rangeOf('LAST_30', today, { from: today, to: today });
}
