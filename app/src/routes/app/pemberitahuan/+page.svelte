<script lang="ts">
	import NotificationList from '#lib/features/notifications/components/NotificationList.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.notifications.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.notifications.title}</h1>
	<NotificationList />
</section>
