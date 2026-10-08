<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import OutboxReview from '#lib/features/outbox/components/OutboxReview.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.outbox.reviewTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.outbox.reviewTitle}</h1>
	<OutboxReview entryId={page.params.id ?? ''} ondiscarded={() => void goto('/app/antrean')} />
</section>
