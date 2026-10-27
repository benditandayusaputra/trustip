import { getContext, setContext, type Component, type Snippet } from 'svelte';

export interface PageTitleEntry {
	title: Snippet;
	icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
	backHref: string | undefined;
}

const key = Symbol('page-title');

export class PageTitle {
	entry = $state.raw<PageTitleEntry>();
	readonly #active: () => boolean;

	constructor(active: () => boolean) {
		this.#active = active;
	}

	get active(): boolean {
		return this.#active();
	}

	show(entry: PageTitleEntry): () => void {
		this.entry = entry;
		return () => {
			if (this.entry === entry) {
				this.entry = undefined;
			}
		};
	}
}

export function setPageTitle(title: PageTitle): void {
	setContext(key, title);
}

export function getPageTitle(): PageTitle | undefined {
	return getContext<PageTitle | undefined>(key);
}
