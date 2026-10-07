<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import NotificationBell from '#lib/features/notifications/components/NotificationBell.svelte';
	import NotificationList from '#lib/features/notifications/components/NotificationList.svelte';

	setToaster(new ToastQueue());
</script>

<NotificationBell />
<NotificationList />
<Toaster />
