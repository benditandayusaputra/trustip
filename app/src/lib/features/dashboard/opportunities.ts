import type {
	DirectoryConsignorCard,
	DirectoryStoreCard,
	NeedBoardCard
} from '#lib/api/directory.js';
import type { Need } from '#lib/api/needs.js';
import { formatDistance } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

export interface Opportunity {
	key: string;
	title: string;
	detail: string;
	href: string;
}

export const opportunitiesPerKind = 3;

const text = id.dashboard;

type PartnerCard = DirectoryStoreCard | DirectoryConsignorCard;

export function isNewPartner(card: PartnerCard): boolean {
	return !card.is_partner && !card.is_self;
}

export function needOpportunity(card: NeedBoardCard): Opportunity {
	const distance = card.distance_m === null ? '' : formatDistance(card.distance_m);
	return {
		key: `need-${card.id}`,
		title: card.title,
		detail: text.nearbyNeed(card.store.name, distance),
		href: `/app/ajuan/baru?${new URLSearchParams({ sasaran: card.store.id, kebutuhan: card.id }).toString()}`
	};
}

export function partnerOpportunity(card: PartnerCard, searchPage: string): Opportunity {
	return {
		key: `partner-${card.id}`,
		title: card.name,
		detail: text.suggestedPartner(card.match?.score ?? 0),
		href: `${searchPage}?${new URLSearchParams([['q', card.name]]).toString()}`
	};
}

export function responseOpportunity(need: Need): Opportunity {
	return {
		key: `response-${need.id}`,
		title: need.title,
		detail: text.needResponses(need.response_count),
		href: `/app/kebutuhan/${need.id}`
	};
}
