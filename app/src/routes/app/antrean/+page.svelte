<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import CloudUpload from '@lucide/svelte/icons/cloud-upload';
	import OutboxList from '#lib/features/outbox/components/OutboxList.svelte';
	import { offline as id } from '#lib/i18n/offline.js';
</script>

<svelte:head>
	<title>{id.outbox.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<PageHeading icon={CloudUpload}>{id.outbox.title}</PageHeading>
	<p class="text-base text-text-muted">{id.outbox.intro}</p>
	<OutboxList />
</section>
