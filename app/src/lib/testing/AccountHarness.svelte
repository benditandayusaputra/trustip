<script lang="ts">
	import { untrack } from 'svelte';
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import Greeting from '#lib/features/account/components/Greeting.svelte';
	import SessionEnd from '#lib/features/account/components/SessionEnd.svelte';
	import type { Placement } from '#lib/features/account/session.svelte.js';
	import { setSessionSlot, type SessionSlot } from '#lib/features/account/session-slot.js';

	interface Props {
		placement?: Placement;
		onsignedout: () => void;
		onmisplaced?: () => void;
		everywhere?: boolean;
		slot?: SessionSlot;
	}

	let {
		placement = 'member',
		onsignedout,
		onmisplaced = () => undefined,
		everywhere = false,
		slot
	}: Props = $props();

	setToaster(new ToastQueue());
	untrack(() => {
		if (slot) {
			setSessionSlot(slot);
		}
	});
</script>

<AccountGate {placement} {onsignedout} {onmisplaced}>
	<Greeting />
	{#if everywhere}<SessionEnd everywhere />{/if}
</AccountGate>
<Toaster />
