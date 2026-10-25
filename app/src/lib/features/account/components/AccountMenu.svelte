<script lang="ts">
	import UserRound from '@lucide/svelte/icons/user-round';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { offline as offlineText } from '#lib/i18n/offline.js';
	import { discardOwnEntries, outboxScheduler, waitingEntries } from '#lib/offline/device.js';
	import { Leaving } from '../leaving.svelte';
	import { getAccountSession } from '../session.svelte';

	const id = { ...commonText, ...offlineText };

	const text = id.account;
	const outboxText = id.outbox;
	const session = getAccountSession();
	const toaster = getToaster();
	const leaving = new Leaving({
		countWaiting: async () => (await waitingEntries()).length,
		sendWaiting: () => outboxScheduler().run(),
		discardWaiting: discardOwnEntries
	});
	let confirming = $state(false);
	let busy = $state(false);

	async function end(action: () => Promise<void>) {
		busy = true;
		try {
			await action();
		} catch (error) {
			toaster.show({ tone: 'danger', message: error instanceof Error ? error.message : '' });
		} finally {
			busy = false;
		}
	}

	async function sendFirst() {
		const left = await leaving.sendFirst();
		if (left > 0) {
			toaster.show({ tone: 'info', message: outboxText.stillWaiting(left) });
		}
	}
</script>

<details class="relative">
	<summary
		aria-label={text.openMenu}
		class="inline-flex min-h-11 min-w-11 cursor-pointer items-center justify-center gap-2 rounded-pill border border-border bg-surface px-2 font-semibold text-text md:px-4"
	>
		<UserRound class="size-4 text-accent" aria-hidden="true" /><span class="sr-only md:not-sr-only"
			>{text.menu}</span
		><ChevronDown class="hidden size-4 text-text-muted md:block" aria-hidden="true" />
	</summary>
	<nav
		aria-label={text.menu}
		class="absolute right-0 z-10 mt-2 flex w-max flex-col gap-2 rounded-card border border-border bg-surface p-2 shadow-raised"
	>
		<a href="/panduan" class="secondary-link">{id.nav.guide}</a>
		<a href="/faq" class="secondary-link">{id.nav.faq}</a>
		<Button
			variant="secondary"
			{busy}
			onclick={() => end(() => leaving.begin(() => session.endThisDevice()))}
		>
			{text.logOut}
		</Button>
		<Button variant="secondary" disabled={busy} onclick={() => (confirming = true)}>
			{text.logOutEverywhere}
		</Button>
	</nav>
</details>

<Dialog bind:open={confirming} title={text.confirmEverywhereTitle}>
	<p class="text-base text-text">{text.confirmEverywhereBody}</p>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (confirming = false)}>{text.cancel}</Button>
		<Button
			variant="danger"
			{busy}
			onclick={() =>
				end(() => {
					confirming = false;
					return leaving.begin(() => session.endEveryDevice());
				})}
		>
			{text.logOutEverywhere}
		</Button>
	{/snippet}
</Dialog>

<Dialog bind:open={leaving.asking} title={outboxText.leaveTitle(leaving.waiting)}>
	<p class="text-base text-text">{outboxText.leaveBody}</p>
	{#snippet actions()}
		<Button variant="danger" disabled={busy} onclick={() => end(() => leaving.discard())}>
			{outboxText.discardAll}
		</Button>
		<Button variant="secondary" disabled={busy} onclick={() => end(() => leaving.keepOnDevice())}>
			{outboxText.keepOnDevice}
		</Button>
		<Button {busy} onclick={() => end(sendFirst)}>{outboxText.sendFirst}</Button>
	{/snippet}
</Dialog>
