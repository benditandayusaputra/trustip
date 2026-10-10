import type { Insight } from '#lib/api/insights.js';
import type { Tone } from '#lib/components/ui/tone.js';
import { formatRupiah } from '#lib/i18n/format.js';
import { dashboard as id } from '#lib/i18n/dashboard.js';

export interface InsightCard {
	key: string;
	tone: Tone;
	level: string;
	title: string;
	body: string;
	actionLabel: string;
	href: string;
}

type Kind = Insight['kind'];
type Action = Insight['action'];

const text = id.insights;

const tones: Record<Insight['level'], Tone> = { HIGH: 'danger', MEDIUM: 'pending', LOW: 'info' };

const titles: Record<Kind, (insight: Insight) => string> = {
	EXPIRED: ({ subject }) => text.titles.EXPIRED(subject.product_name),
	NEAR_EXPIRY: ({ subject }) => text.titles.NEAR_EXPIRY(subject.product_name),
	LOW_STOCK: ({ subject }) => text.titles.LOW_STOCK(subject.product_name),
	SETTLEMENT_OVERDUE: ({ subject }) => text.titles.SETTLEMENT_OVERDUE(subject.settlement_number),
	NOT_COUNTED: ({ subject }) => text.titles.NOT_COUNTED(subject.counterpart_name),
	MATCHING_PARTNER: ({ subject }) => text.titles.MATCHING_PARTNER(subject.organization_name)
};

const bodies: Record<Kind, (insight: Insight) => string> = {
	EXPIRED: ({ subject, units }) =>
		text.bodies.EXPIRED(subject.counterpart_name, subject.batch_code, units),
	NEAR_EXPIRY: ({ subject, days, units }) =>
		text.bodies.NEAR_EXPIRY(subject.counterpart_name, days, units),
	LOW_STOCK: ({ subject, days }) => text.bodies.LOW_STOCK(subject.counterpart_name, days),
	SETTLEMENT_OVERDUE: ({ subject, days, amount }) =>
		text.bodies.SETTLEMENT_OVERDUE(subject.counterpart_name, days, formatRupiah(amount)),
	NOT_COUNTED: ({ days, units }) => text.bodies.NOT_COUNTED(days, units),
	MATCHING_PARTNER: ({ score }) => text.bodies.MATCHING_PARTNER(score)
};

function partnerPage({ subject }: Insight): string {
	return `/app/mitra/${subject.partnership_id ?? ''}`;
}

function settlementPage({ subject }: Insight): string {
	return `/app/setoran/${subject.settlement_id ?? ''}`;
}

function countForm({ subject }: Insight): string {
	return `/app/hitung/baru?mitra=${subject.partnership_id ?? ''}`;
}

const targets: Record<Action, (insight: Insight) => string> = {
	COUNT_WITH_RETURNS: (insight) =>
		`${countForm(insight)}&produk=${insight.subject.product_id ?? ''}&kembali=${String(insight.units)}`,
	COUNT: countForm,
	REMIND_STORE: partnerPage,
	VIEW_DROP_SUGGESTION: partnerPage,
	ASK_RESTOCK: partnerPage,
	REMIND_PAYMENT: settlementPage,
	RECORD_PAYMENT: settlementPage,
	REMIND_COUNT: partnerPage,
	OPEN_PROFILE: ({ subject }) => `?q=${encodeURIComponent(subject.organization_name)}`
};

const searchPages: Record<Side, string> = { CONSIGNOR: '/cari/toko', STORE: '/cari/penitip' };

export type Side = 'CONSIGNOR' | 'STORE';

function hrefOf(insight: Insight, side: Side): string {
	const target = targets[insight.action](insight);
	return insight.action === 'OPEN_PROFILE' ? searchPages[side] + target : target;
}

export function cardOf(insight: Insight, index: number, side: Side): InsightCard {
	return {
		key: `${insight.kind}-${String(index)}`,
		tone: tones[insight.level],
		level: id.dashboard.levels[insight.level],
		title: titles[insight.kind](insight),
		body: bodies[insight.kind](insight),
		actionLabel: text.actions[insight.action],
		href: hrefOf(insight, side)
	};
}
