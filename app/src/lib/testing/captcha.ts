import type { TurnstileApi, TurnstileRenderOptions } from '#lib/components/ui/turnstile.js';

export interface FakeCaptcha {
	solve: (token: string) => void;
	resets: () => number;
}

export function installFakeCaptcha(): FakeCaptcha {
	let options: TurnstileRenderOptions | undefined;
	let resetCount = 0;
	const api: TurnstileApi = {
		render: (_container: HTMLElement, rendered: TurnstileRenderOptions) => {
			options = rendered;
			return 'widget-1';
		},
		reset: () => {
			resetCount += 1;
		},
		remove: () => undefined
	};
	window.turnstile = api;
	return {
		solve: (token) => {
			if (!options) {
				throw new Error('captcha widget was not rendered yet');
			}
			options.callback(token);
		},
		resets: () => resetCount
	};
}
