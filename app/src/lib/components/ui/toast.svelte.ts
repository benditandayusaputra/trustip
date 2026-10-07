import { createContext } from 'svelte';
import type { Tone } from './tone';

interface ToastAction {
	label: string;
	href?: string;
	onclick?: () => void;
}

export interface ToastInput {
	tone: Tone;
	message: string;
	action?: ToastAction;
}

export interface Toast extends ToastInput {
	id: number;
}

const maxVisibleToasts = 3;

export class Toaster {
	items = $state<Toast[]>([]);
	#lastId = 0;

	show(input: ToastInput): number {
		this.#lastId += 1;
		const toast: Toast = { ...input, id: this.#lastId };
		this.items = [...this.items, toast].slice(-maxVisibleToasts);
		return toast.id;
	}

	dismiss(toastId: number): void {
		this.items = this.items.filter((toast) => toast.id !== toastId);
	}
}

export const [getToaster, setToaster] = createContext<Toaster>();
