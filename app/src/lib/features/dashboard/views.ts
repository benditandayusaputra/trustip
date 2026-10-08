import type { Dashboard } from '#lib/api/insights.js';
import type { SalesNarrowing, SalesReport } from '#lib/api/reports.js';
import { rangeOf, type DateRange } from '#lib/domain/ranges.js';
import { businessDateOf, formatRupiah } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';
import type { RankedRow } from './components/RankingList.svelte';
import type { TaskLink } from './components/TaskList.svelte';

const maxRankedRows = 5;
const text = id.dashboard;

type Tasks = Dashboard['tasks'];

interface TaskSpec {
	count: (tasks: Tasks) => number;
	label: (count: number) => string;
	href: string;
}

export const consignorTasks: TaskSpec[] = [
	{
		count: (tasks) => tasks.drops_to_confirm,
		label: text.tasks.dropsToConfirm,
		href: '/app/titipan'
	},
	{
		count: (tasks) => tasks.counts_to_confirm,
		label: text.tasks.countsToConfirm,
		href: '/app/hitung'
	},
	{
		count: (tasks) => tasks.payments_to_review,
		label: text.tasks.paymentsToReview,
		href: '/app/setoran'
	},
	{
		count: (tasks) => tasks.proposals_awaiting,
		label: text.tasks.proposalsAwaiting,
		href: '/app/ajuan'
	}
];

export function taskLinksOf(tasks: Tasks, specs: TaskSpec[]): TaskLink[] {
	return specs
		.map((spec) => ({ count: spec.count(tasks), label: spec.label, href: spec.href }))
		.filter((task) => task.count > 0)
		.map((task) => ({ label: task.label(task.count), href: task.href }));
}

export function rankedByGross(report: SalesReport | undefined): RankedRow[] {
	return (report?.rows ?? [])
		.filter((row) => row.sold_qty > 0)
		.slice(0, maxRankedRows)
		.map((row) => ({ key: row.key, label: row.label, value: formatRupiah(row.split.gross) }));
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
