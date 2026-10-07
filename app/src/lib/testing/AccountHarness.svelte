<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import AccountWelcome from '#lib/features/account/components/AccountWelcome.svelte';

	interface Props {
		onsignedout: () => void;
	}

	let { onsignedout }: Props = $props();

	setToaster(new ToastQueue());
</script>

<AccountGate {onsignedout}>
	<AccountWelcome />
</AccountGate>
<Toaster />
