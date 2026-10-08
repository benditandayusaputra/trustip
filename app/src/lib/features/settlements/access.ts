import type { Settlement } from '#lib/api/settlements.js';

export interface Viewer {
	organizationId: string;
	isOwner: boolean;
}

function hasOpenPayment(settlement: Settlement): boolean {
	return settlement.payments.some((payment) => payment.status !== 'REJECTED');
}

function isOpen(settlement: Settlement): boolean {
	return settlement.status === 'ISSUED' && !hasOpenPayment(settlement);
}

export function canRecordPayment(settlement: Settlement, viewer: Viewer): boolean {
	return viewer.isOwner && viewer.organizationId === settlement.store.id && isOpen(settlement);
}

export function canMarkPaid(settlement: Settlement, viewer: Viewer): boolean {
	return viewer.isOwner && viewer.organizationId === settlement.consignor.id && isOpen(settlement);
}

export function canReview(settlement: Settlement, viewer: Viewer): boolean {
	const isConsignorOwner = viewer.isOwner && viewer.organizationId === settlement.consignor.id;
	return (
		isConsignorOwner &&
		settlement.payments.some((payment) => payment.status === 'PENDING_CONFIRMATION')
	);
}

export function canRemind(settlement: Settlement, viewer: Viewer): boolean {
	return viewer.organizationId === settlement.consignor.id && settlement.status === 'ISSUED';
}
