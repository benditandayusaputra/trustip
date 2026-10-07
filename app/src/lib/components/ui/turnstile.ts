const scriptUrl = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';

export interface TurnstileRenderOptions {
	sitekey: string;
	action: string;
	theme: 'dark';
	language: string;
	callback: (token: string) => void;
	'expired-callback': () => void;
	'error-callback': () => void;
}

export interface TurnstileApi {
	render: (container: HTMLElement, options: TurnstileRenderOptions) => string;
	reset: (widgetId: string) => void;
	remove: (widgetId: string) => void;
}

declare global {
	interface Window {
		turnstile?: TurnstileApi;
	}
}

let loading: Promise<TurnstileApi> | undefined;

export function loadTurnstile(): Promise<TurnstileApi> {
	if (window.turnstile) {
		return Promise.resolve(window.turnstile);
	}
	loading ??= injectScript().catch((error: unknown) => {
		loading = undefined;
		throw error;
	});
	return loading;
}

function injectScript(): Promise<TurnstileApi> {
	return new Promise((resolve, reject) => {
		const script = document.createElement('script');
		script.src = scriptUrl;
		script.async = true;
		script.addEventListener('load', () => {
			if (window.turnstile) {
				resolve(window.turnstile);
				return;
			}
			reject(new Error('Turnstile script loaded without its API'));
		});
		script.addEventListener('error', () => {
			reject(new Error('Turnstile script failed to load'));
		});
		document.head.append(script);
	});
}
