import {
	actionDoneEvent,
	dismissInstall,
	hasCompletedAction,
	isInstallDismissed
} from '#lib/offline/install.js';

interface InstallPromptEvent extends Event {
	prompt: () => Promise<void>;
	userChoice: Promise<{ outcome: string }>;
}

export type InviteKind = 'hidden' | 'prompt' | 'ios';

export interface InstallEnvironment {
	isInstalled: () => boolean;
	isIos: () => boolean;
}

const iosDevices = /iphone|ipad|ipod/iu;

export const browserEnvironment: InstallEnvironment = {
	isInstalled: () => window.matchMedia('(display-mode: standalone)').matches,
	isIos: () => iosDevices.test(navigator.userAgent)
};

export class InstallInvite {
	kind = $state<InviteKind>('hidden');
	#deferred: InstallPromptEvent | undefined;
	readonly #environment: InstallEnvironment;

	constructor(environment: InstallEnvironment) {
		this.#environment = environment;
	}

	watch(): () => void {
		const keep = (event: Event) => {
			event.preventDefault();
			this.#deferred = event as InstallPromptEvent;
			this.#decide();
		};
		const decide = () => {
			this.#decide();
		};
		window.addEventListener('beforeinstallprompt', keep);
		window.addEventListener(actionDoneEvent, decide);
		this.#decide();
		return () => {
			window.removeEventListener('beforeinstallprompt', keep);
			window.removeEventListener(actionDoneEvent, decide);
		};
	}

	async install(): Promise<void> {
		const deferred = this.#deferred;
		this.dismiss();
		if (deferred) {
			await deferred.prompt();
			await deferred.userChoice;
		}
	}

	dismiss(): void {
		dismissInstall();
		this.#deferred = undefined;
		this.kind = 'hidden';
	}

	#decide(): void {
		this.kind = this.#isWelcome() ? this.#inviteKind() : 'hidden';
	}

	#isWelcome(): boolean {
		return hasCompletedAction() && !isInstallDismissed() && !this.#environment.isInstalled();
	}

	#inviteKind(): InviteKind {
		if (this.#deferred) {
			return 'prompt';
		}
		return this.#environment.isIos() ? 'ios' : 'hidden';
	}
}
