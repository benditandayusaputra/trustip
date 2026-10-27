<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Bell from '@lucide/svelte/icons/bell';
	import NotificationList from '#lib/features/notifications/components/NotificationList.svelte';
	import { common as id } from '#lib/i18n/common.js';
</script>

<svelte:head>
	<title>{id.notifications.title}</title>
</svelte:head>

<section class="flex w-full flex-col gap-6">
	<PageHeading icon={Bell}>{id.notifications.title}</PageHeading>
	<NotificationList />
</section>
