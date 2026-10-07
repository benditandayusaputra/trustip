<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import OnboardingWizard from '#lib/features/organization/components/OnboardingWizard.svelte';

	interface Props {
		onfounded: () => void;
	}

	let { onfounded }: Props = $props();

	setToaster(new ToastQueue());
</script>

<OnboardingWizard {onfounded} />
<Toaster />
