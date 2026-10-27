<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { offline as offlineText } from '#lib/i18n/offline.js';
	import { discardOwnEntries, outboxScheduler, waitingEntries } from '#lib/offline/device.js';
	import { Leaving } from '../leaving.svelte';
	import { getAccountSession } from '../session.svelte';

	let { everywhere = false }: { everywhere?: boolean } = $props();

	const text = commonText.account;
	const outboxText = offlineText.outbox;
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

{#if everywhere}
	<Button variant="secondary" disabled={busy} onclick={() => (confirming = true)}>
		{text.logOutEverywhere}
	</Button>
{:else}
	<Button
		variant="secondary"
		{busy}
		onclick={() => end(() => leaving.begin(() => session.endThisDevice()))}
	>
		{text.logOut}
	</Button>
{/if}

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
