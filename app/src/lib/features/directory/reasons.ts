import type { MatchReason } from '#lib/api/directory.js';
import { formatBasisPoints, formatDistance } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

const text = id.directory.reasons;

const formatters: Record<MatchReason['code'], (reason: MatchReason) => string> = {
	DISTANCE: (reason) => formatDistance(reason.distance_m ?? 0),
	NEEDS_CATEGORY: (reason) => text.needsCategory(reason.category ?? ''),
	ACCEPTS_CATEGORY: (reason) => text.acceptsCategory(reason.category ?? ''),
	ON_TIME_PAYMENT: (reason) => id.directory.onTime(formatBasisPoints(reason.on_time_bps ?? 0)),
	ESTABLISHED: () => text.established,
	PREFERRED_STORE_TYPE: () => text.preferredStoreType,
	ACTIVE_RECENTLY: () => text.activeRecently
};

export function reasonText(reason: MatchReason, isApproximate = false): string {
	const formatted = formatters[reason.code](reason);
	return isApproximate && reason.code === 'DISTANCE'
		? id.consignorDirectory.approximate(formatted)
		: formatted;
}
