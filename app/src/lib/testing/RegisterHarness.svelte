<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import RegisterForm from '#lib/features/auth/components/RegisterForm.svelte';

	interface Props {
		onregistered: () => void;
	}

	let { onregistered }: Props = $props();

	setToaster(new ToastQueue());
</script>

<RegisterForm siteKey="site-key" {onregistered} />
<Toaster />
