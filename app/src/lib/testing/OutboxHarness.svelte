<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import OutboxBadge from '#lib/features/outbox/components/OutboxBadge.svelte';
	import OutboxList from '#lib/features/outbox/components/OutboxList.svelte';

	setToaster(new ToastQueue());
</script>

<OutboxBadge />
<OutboxList />
<Toaster />
