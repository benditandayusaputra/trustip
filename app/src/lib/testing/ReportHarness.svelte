<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import ReportListing from '#lib/features/directory/components/ReportListing.svelte';

	setToaster(new ToastQueue());
</script>

<ReportListing
	target={{ target_type: 'PRODUCT', target_id: '0192f0c4-0000-7000-8000-000000000101' }}
	name="Roti Sobek"
/>
<Toaster />
