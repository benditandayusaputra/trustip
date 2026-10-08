import { describe, expect, it, vi } from 'vitest';
import type { EntryDraft } from './outbox';
import { OutboxScheduler } from './scheduler';
import type { FlushResult, OutboxSync, SubmitResult } from './sync';

function fakeSync(results: FlushResult[]) {
	return {
		enqueue: vi.fn(() => Promise.resolve('entry-1')),
		flush: vi.fn(() => Promise.resolve(results.shift() ?? { kind: 'done', sent: 0 })),
		resultOf: vi.fn((): Promise<SubmitResult> => Promise.resolve({ kind: 'queued' }))
	};
}

function schedulerWith(results: FlushResult[]) {
	const sync = fakeSync(results);
	const timers: { task: () => void; delayMs: number }[] = [];
	const afterSent = vi.fn();
	const scheduler = new OutboxScheduler({
		sync: sync as unknown as OutboxSync,
		afterSent,
		timers: {
			set: (task, delayMs) => timers.push({ task, delayMs }),
			clear: vi.fn()
		},
		random: () => 0.5
	});
	return { scheduler, sync, timers, afterSent };
}

const draft = { kind: 'COUNT' } as EntryDraft;

describe('OutboxScheduler (R-OFF-05)', () => {
	it('tries again later with a growing pause after a temporary failure', async () => {
		const { scheduler, timers } = schedulerWith([{ kind: 'retryLater', attempts: 2 }]);

		await scheduler.run();

		expect(timers.map((timer) => timer.delayMs)).toEqual([8_000]);
	});

	it('refreshes the work packets after something was sent', async () => {
		const { scheduler, afterSent } = schedulerWith([{ kind: 'done', sent: 1 }]);

		await scheduler.run();

		expect(afterSent).toHaveBeenCalledOnce();
	});

	it('R-OFF-02 saves a form to the queue first and then sends at once', async () => {
		const { scheduler, sync } = schedulerWith([]);

		await expect(scheduler.submit(draft)).resolves.toEqual({ kind: 'queued' });

		expect(sync.enqueue).toHaveBeenCalledWith(draft);
		expect(sync.flush).toHaveBeenCalledOnce();
		expect(sync.resultOf).toHaveBeenCalledWith('entry-1');
	});
});
