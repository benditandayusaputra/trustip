<script lang="ts">
	import { untrack } from 'svelte';
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import type { AccountProfile } from '#lib/api/account.js';
	import AccountCard from '#lib/features/account/components/AccountCard.svelte';

	let { account: initial }: { account: AccountProfile } = $props();

	setToaster(new ToastQueue());
	let account = $state(untrack(() => initial));
</script>

<AccountCard
	{account}
	onsaved={(saved: AccountProfile) => {
		account = saved;
	}}
/>
<Toaster />
