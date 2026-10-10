<script lang="ts">
	import OutboxList from '#lib/features/outbox/components/OutboxList.svelte';
	import { offline as id } from '#lib/i18n/offline.js';
</script>

<svelte:head>
	<title>{id.outbox.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.outbox.title}</h1>
	<p class="text-base text-text-muted">{id.outbox.intro}</p>
	<OutboxList />
</section>
