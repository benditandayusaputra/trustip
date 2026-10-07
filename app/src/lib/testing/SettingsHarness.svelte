<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import OrganizationSettings from '#lib/features/organization/components/OrganizationSettings.svelte';

	interface Props {
		organizationId: string;
		canEdit: boolean;
	}

	let { organizationId, canEdit }: Props = $props();

	setToaster(new ToastQueue());
</script>

<OrganizationSettings {organizationId} {canEdit} />
<Toaster />
