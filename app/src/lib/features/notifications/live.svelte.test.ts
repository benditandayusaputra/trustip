import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { StreamHandlers } from '#lib/api/stream.js';
import {
	hiddenGraceMs,
	LiveNotifications,
	pollIntervalMs,
	settleBeforeConnectMs,
	type LiveDeps
} from './live.svelte';

interface Fixture {
	live: LiveNotifications;
	refresh: ReturnType<typeof vi.fn>;
	close: ReturnType<typeof vi.fn>;
	connect: ReturnType<typeof vi.fn>;
	handlers: () => StreamHandlers;
}

function fixture(connects: boolean, visible = true): Fixture {
	let captured: StreamHandlers | null = null;
	const close = vi.fn();
	const refresh = vi.fn(() => Promise.resolve());
	const connect = vi.fn((handlers: StreamHandlers) => {
		captured = handlers;
		return connects ? Promise.resolve(close) : Promise.reject(new Error('down'));
	});
	const deps: LiveDeps = { connect, refresh, onarrive: vi.fn(), isVisible: () => visible };
	return {
		live: new LiveNotifications(deps),
		refresh,
		close,
		connect,
		handlers: () => {
			if (captured === null) {
				throw new Error('not connected');
			}
			return captured;
		}
	};
}

beforeEach(() => {
	vi.useFakeTimers();
});

afterEach(() => {
	vi.useRealTimers();
});

describe('LiveNotifications', () => {
	it('waits for the page to settle before asking for a stream ticket', async () => {
		const { live, connect } = fixture(true);

		live.start();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs - 1);
		live.stop();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);

		expect(connect).not.toHaveBeenCalled();
	});

	it('FR-NTF-02 stays live while the stream is open', async () => {
		const { live, refresh } = fixture(true);

		live.start();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);
		await vi.advanceTimersByTimeAsync(pollIntervalMs * 2);

		expect(live.mode).toBe('live');
		expect(refresh).not.toHaveBeenCalled();
	});

	it('FR-NTF-03 checks every 30 seconds while visible after the stream fails', async () => {
		const { live, refresh } = fixture(false);

		live.start();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);
		await vi.advanceTimersByTimeAsync(pollIntervalMs * 2);

		expect(live.mode).toBe('polling');
		expect(refresh).toHaveBeenCalledTimes(2);
	});

	it('FR-NTF-03 does not check while the tab is hidden', async () => {
		const { live, refresh } = fixture(false, false);

		live.start();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);
		await vi.advanceTimersByTimeAsync(pollIntervalMs * 2);

		expect(refresh).not.toHaveBeenCalled();
	});

	it('FR-NTF-03 checks once and reconnects when an open stream drops', async () => {
		const { live, refresh, close, connect, handlers } = fixture(true);
		live.start();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);

		handlers().onfailure();
		expect(live.mode).toBe('polling');
		await vi.advanceTimersByTimeAsync(pollIntervalMs);

		expect(close).toHaveBeenCalled();
		expect(refresh).toHaveBeenCalledTimes(1);
		expect(connect).toHaveBeenCalledTimes(2);
		expect(live.mode).toBe('live');
		await vi.advanceTimersByTimeAsync(pollIntervalMs * 2);
		expect(refresh).toHaveBeenCalledTimes(1);
	});

	it('BR-NTF-05 disconnects after five hidden minutes and reconnects when visible', async () => {
		const { live, close, connect, refresh } = fixture(true);
		live.start();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);

		live.hide();
		await vi.advanceTimersByTimeAsync(hiddenGraceMs);
		expect(live.mode).toBe('paused');
		expect(close).toHaveBeenCalled();

		live.show();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);
		expect(live.mode).toBe('live');
		expect(connect).toHaveBeenCalledTimes(2);
		expect(refresh).toHaveBeenCalledTimes(1);
	});

	it('BR-NTF-05 keeps the stream when the tab returns within five minutes', async () => {
		const { live, close } = fixture(true);
		live.start();
		await vi.advanceTimersByTimeAsync(settleBeforeConnectMs);

		live.hide();
		await vi.advanceTimersByTimeAsync(hiddenGraceMs - 1);
		live.show();
		await vi.advanceTimersByTimeAsync(hiddenGraceMs);

		expect(live.mode).toBe('live');
		expect(close).not.toHaveBeenCalled();
	});
});
