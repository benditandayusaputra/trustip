<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import LoginForm from '#lib/features/auth/components/LoginForm.svelte';

	interface Props {
		onsignedin: () => void;
	}

	let { onsignedin }: Props = $props();

	setToaster(new ToastQueue());
</script>

<LoginForm siteKey="site-key" {onsignedin} />
<Toaster />
