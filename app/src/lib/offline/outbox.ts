import type { CountRequest } from '#lib/api/counts.js';
import type { DropRequest } from '#lib/api/drops.js';
import { ApiError, NetworkError, type FieldError } from '#lib/api/errors.js';
import type { Terms } from '#lib/domain/pricing.js';

type OutboxStatus = 'PENDING' | 'SENDING' | 'SENT' | 'HELD' | 'NEEDS_REVIEW';

export interface OutboxProblem {
	status: number;
	code: string;
	message: string;
	requestId: string;
	fields: FieldError[];
}

export interface QueuedDropLine {
	productId: string;
	name: string;
	unit: string;
	qty: number;
	expiryDate: string;
	terms: Terms;
}

export interface SentDocument {
	id: string;
	number: string;
}

interface EntryBase {
	id: string;
	revision: number;
	idempotencyKey: string;
	partnershipId: string;
	partnerName: string;
	recordedAt: string;
	revisedAt?: string;
	savedOffline: boolean;
	userId: string;
	organizationId: string;
	status: OutboxStatus;
	attempts: number;
	problem: OutboxProblem | null;
	document: SentDocument | null;
}

export type OutboxEntry =
	| (EntryBase & { kind: 'DROP'; request: DropRequest; lines: QueuedDropLine[] })
	| (EntryBase & { kind: 'COUNT'; request: CountRequest });

export type EntryDraft =
	| {
			kind: 'DROP';
			partnershipId: string;
			partnerName: string;
			request: DropRequest;
			lines: QueuedDropLine[];
	  }
	| { kind: 'COUNT'; partnershipId: string; partnerName: string; request: CountRequest };

export type SendVerdict = 'retry' | 'review' | 'signedOut';

const unauthorizedStatus = 401;
const timeoutStatus = 408;
const tooManyStatus = 429;
const firstServerStatus = 500;
const retryBaseMs = 2_000;
const backoffGrowth = 2;
const spreadFloor = 0.5;
const retryCapMs = 300_000;
const oldEntryMs = 86_400_000;

export function isWaiting(entry: OutboxEntry): boolean {
	return entry.status !== 'SENT';
}

function ageMs(entry: OutboxEntry, now: Date): number {
	return now.getTime() - Date.parse(entry.recordedAt);
}

export function isOld(entry: OutboxEntry, now: Date): boolean {
	return isWaiting(entry) && ageMs(entry, now) > oldEntryMs;
}

export function isDeliveredLongAgo(entry: OutboxEntry, now: Date): boolean {
	return !isWaiting(entry) && ageMs(entry, now) > oldEntryMs;
}

export function verdictOf(error: unknown): SendVerdict {
	if (error instanceof NetworkError || !(error instanceof ApiError)) {
		return 'retry';
	}
	if (error.isSessionProblem() || error.status === unauthorizedStatus) {
		return 'signedOut';
	}
	const isTemporary =
		error.status === timeoutStatus ||
		error.status === tooManyStatus ||
		error.status >= firstServerStatus;
	return isTemporary ? 'retry' : 'review';
}

const networkProblem = 'NETWORK';

export function problemOf(error: unknown): OutboxProblem {
	if (!(error instanceof ApiError)) {
		const message = error instanceof Error ? error.message : '';
		return { status: 0, code: networkProblem, message, requestId: '', fields: [] };
	}
	return {
		status: error.status,
		code: error.code,
		message: error.message,
		requestId: error.requestId,
		fields: error.fields
	};
}

export function errorOf(problem: OutboxProblem): ApiError {
	return new ApiError({ ...problem, retryAfterSeconds: undefined });
}

function byRecording(left: OutboxEntry, right: OutboxEntry): number {
	return left.recordedAt.localeCompare(right.recordedAt) || left.id.localeCompare(right.id);
}

export function sendingQueues(entries: OutboxEntry[]): OutboxEntry[][] {
	const queues = new Map<string, OutboxEntry[]>();
	for (const entry of entries.filter(isWaiting).sort(byRecording)) {
		queues.set(entry.partnershipId, [...(queues.get(entry.partnershipId) ?? []), entry]);
	}
	return [...queues.values()];
}

export function retryDelayMs(attempts: number, random: number): number {
	const backoff = Math.min(retryBaseMs * backoffGrowth ** attempts, retryCapMs);
	return Math.round(backoff * (spreadFloor + random));
}
