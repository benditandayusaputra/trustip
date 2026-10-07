<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import EmptyDashboard from '#lib/features/account/components/EmptyDashboard.svelte';
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
	<EmptyDashboard />
</AccountGate>
<Toaster />
