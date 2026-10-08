import { describe, expect, it } from 'vitest';
import { ApiError, NetworkError } from '#lib/api/errors.js';
import { isOld, retryDelayMs, sendingQueues, verdictOf, type OutboxEntry } from './outbox';

const apiError = (status: number, code: string) =>
	new ApiError({
		status,
		code,
		message: code,
		requestId: 'r',
		fields: [],
		retryAfterSeconds: undefined
	});

function anEntry(changes: Partial<OutboxEntry>): OutboxEntry {
	return {
		id: 'e1',
		revision: 0,
		idempotencyKey: 'e1',
		kind: 'COUNT',
		partnershipId: 'p1',
		partnerName: 'Dapur Bu Rina',
		request: {
			partnership_id: 'p1',
			agreement_version: 1,
			counted_at: null,
			occasion: 'STORE_CLOSING',
			note: '',
			lines: []
		},
		recordedAt: '2026-11-10T02:00:00.000Z',
		savedOffline: true,
		userId: 'u1',
		organizationId: 'o1',
		status: 'PENDING',
		attempts: 0,
		problem: null,
		document: null,
		...changes
	} as OutboxEntry;
}

describe('verdictOf (R-OFF-09)', () => {
	it('retries a lost connection, a busy server, and a rate limit', () => {
		expect(verdictOf(new NetworkError('putus'))).toBe('retry');
		expect(verdictOf(apiError(503, 'SERVICE_UNAVAILABLE'))).toBe('retry');
		expect(verdictOf(apiError(429, 'RATE_LIMITED'))).toBe('retry');
	});

	it('sends a conflict or a refused rule to review', () => {
		expect(verdictOf(apiError(409, 'STOCK_CHANGED'))).toBe('review');
		expect(verdictOf(apiError(409, 'AGREEMENT_CHANGED'))).toBe('review');
		expect(verdictOf(apiError(422, 'PARTNERSHIP_NOT_ACTIVE'))).toBe('review');
	});

	it('BR-PWA-05 stops on an ended session', () => {
		expect(verdictOf(apiError(401, 'SESSION_EXPIRED'))).toBe('signedOut');
	});
});

describe('sendingQueues (BR-PWA-04)', () => {
	it('orders each partner by recording time and leaves sent entries out', () => {
		const queues = sendingQueues([
			anEntry({ id: 'late', recordedAt: '2026-11-10T03:00:00.000Z' }),
			anEntry({ id: 'other', partnershipId: 'p2' }),
			anEntry({ id: 'early', recordedAt: '2026-11-10T01:00:00.000Z' }),
			anEntry({ id: 'done', status: 'SENT' })
		]);

		expect(queues.map((queue) => queue.map((entry) => entry.id))).toEqual([
			['early', 'late'],
			['other']
		]);
	});
});

describe('retryDelayMs (R-OFF-05)', () => {
	it('doubles the pause after each attempt and stops growing at five minutes', () => {
		expect(retryDelayMs(0, 0.5)).toBe(2_000);
		expect(retryDelayMs(3, 0.5)).toBe(16_000);
		expect(retryDelayMs(20, 0.5)).toBe(300_000);
	});

	it('spreads the pause by up to half either way', () => {
		expect(retryDelayMs(1, 0)).toBe(2_000);
		expect(retryDelayMs(1, 1)).toBe(6_000);
	});
});

describe('isOld (R-OFF-15)', () => {
	it('warns about an entry waiting longer than 24 hours', () => {
		const entry = anEntry({ recordedAt: '2026-11-09T02:00:00.000Z' });

		expect(isOld(entry, new Date('2026-11-10T02:00:00.000Z'))).toBe(false);
		expect(isOld(entry, new Date('2026-11-10T02:00:01.000Z'))).toBe(true);
		expect(isOld({ ...entry, status: 'SENT' }, new Date('2026-11-12T00:00:00.000Z'))).toBe(false);
	});
});
