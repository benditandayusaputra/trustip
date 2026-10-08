export class Freshness {
	staleSince = $state<string | null>(null);

	markStale(storedAt: string): void {
		if (this.staleSince === null || storedAt < this.staleSince) {
			this.staleSince = storedAt;
		}
	}

	markFresh(): void {
		this.staleSince = null;
	}
}
