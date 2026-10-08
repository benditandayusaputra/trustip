<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import Greeting from '#lib/features/account/components/Greeting.svelte';
	import type { Placement } from '#lib/features/account/session.svelte.js';

	interface Props {
		placement?: Placement;
		onsignedout: () => void;
		onmisplaced?: () => void;
	}

	let { placement = 'member', onsignedout, onmisplaced = () => undefined }: Props = $props();

	setToaster(new ToastQueue());
</script>

<AccountGate {placement} {onsignedout} {onmisplaced}>
	<Greeting />
</AccountGate>
<Toaster />
