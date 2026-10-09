import type { LiveNotification, StreamHandlers } from '#lib/api/stream.js';

export type LiveMode = 'idle' | 'live' | 'polling' | 'paused';

export const pollIntervalMs = 30_000;
export const hiddenGraceMs = 300_000;
export const settleBeforeConnectMs = 2000;

export interface LiveDeps {
	connect: (handlers: StreamHandlers) => Promise<() => void>;
	refresh: () => Promise<void>;
	onarrive: (notification: LiveNotification) => void;
	isVisible: () => boolean;
}

export class LiveNotifications {
	mode = $state<LiveMode>('idle');
	readonly #deps: LiveDeps;
	#closeStream: (() => void) | null = null;
	#poll: ReturnType<typeof setInterval> | undefined;
	#pause: ReturnType<typeof setTimeout> | undefined;
	#settle: ReturnType<typeof setTimeout> | undefined;

	constructor(deps: LiveDeps) {
		this.#deps = deps;
	}

	start(): void {
		clearTimeout(this.#settle);
		this.#settle = setTimeout(() => {
			void this.#connect();
		}, settleBeforeConnectMs);
	}

	async #connect(): Promise<void> {
		try {
			this.#closeStream = await this.#deps.connect({
				onnotification: this.#deps.onarrive,
				onfailure: () => {
					this.#fallBack();
				}
			});
			this.mode = 'live';
			this.#stopPolling();
		} catch {
			this.#fallBack();
		}
	}

	hide(): void {
		clearTimeout(this.#pause);
		this.#pause = setTimeout(() => {
			this.#disconnect();
			this.mode = 'paused';
		}, hiddenGraceMs);
	}

	show(): void {
		clearTimeout(this.#pause);
		if (this.mode === 'paused' || this.mode === 'polling') {
			void this.#deps.refresh();
			this.start();
		}
	}

	stop(): void {
		clearTimeout(this.#settle);
		clearTimeout(this.#pause);
		this.#disconnect();
		this.mode = 'idle';
	}

	#fallBack(): void {
		this.#closeStream?.();
		this.#closeStream = null;
		this.mode = 'polling';
		this.#poll ??= setInterval(() => {
			if (this.#deps.isVisible()) {
				void this.#deps.refresh();
			}
		}, pollIntervalMs);
	}

	#disconnect(): void {
		this.#closeStream?.();
		this.#closeStream = null;
		this.#stopPolling();
	}

	#stopPolling(): void {
		clearInterval(this.#poll);
		this.#poll = undefined;
	}
}

export function watchVisibility(live: LiveNotifications): () => void {
	const onvisibilitychange = () => {
		if (document.visibilityState === 'visible') {
			live.show();
			return;
		}
		live.hide();
	};
	const onfocus = () => {
		live.show();
	};
	document.addEventListener('visibilitychange', onvisibilitychange);
	window.addEventListener('focus', onfocus);
	return () => {
		document.removeEventListener('visibilitychange', onvisibilitychange);
		window.removeEventListener('focus', onfocus);
	};
}
