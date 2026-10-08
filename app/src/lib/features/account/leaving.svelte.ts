interface LeavingDeps {
	countWaiting: () => Promise<number>;
	sendWaiting: () => Promise<unknown>;
	discardWaiting: () => Promise<void>;
}

type SignOut = () => Promise<void>;

export class Leaving {
	asking = $state(false);
	waiting = $state(0);
	#signOut: SignOut | undefined;
	readonly #deps: LeavingDeps;

	constructor(deps: LeavingDeps) {
		this.#deps = deps;
	}

	async begin(signOut: SignOut): Promise<void> {
		const waiting = await this.#deps.countWaiting();
		if (waiting === 0) {
			await signOut();
			return;
		}
		this.waiting = waiting;
		this.#signOut = signOut;
		this.asking = true;
	}

	async sendFirst(): Promise<number> {
		await this.#deps.sendWaiting();
		const left = await this.#deps.countWaiting();
		if (left === 0) {
			await this.#finish();
		}
		this.waiting = left;
		return left;
	}

	async keepOnDevice(): Promise<void> {
		await this.#finish();
	}

	async discard(): Promise<void> {
		await this.#deps.discardWaiting();
		await this.#finish();
	}

	async #finish(): Promise<void> {
		this.asking = false;
		await this.#signOut?.();
	}
}
