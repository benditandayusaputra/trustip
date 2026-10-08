<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Count } from '#lib/api/counts.js';
	import CountForm from '#lib/features/counts/components/CountForm.svelte';
	import PartnerChooser from '#lib/features/drops/components/PartnerChooser.svelte';
	import { id } from '#lib/i18n/id.js';

	const partnershipId = $derived(page.url.searchParams.get('mitra') ?? '');
</script>

<svelte:head>
	<title>{id.counts.newTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.counts.newTitle}</h1>
	<p class="text-base text-text-muted">{id.counts.newIntro}</p>
	{#if partnershipId === ''}
		<PartnerChooser hrefOf={(partnerId: string) => `/app/hitung/baru?mitra=${partnerId}`} />
	{:else}
		{#key partnershipId}
			<CountForm
				{partnershipId}
				dropHref="/app/titipan/baru?mitra={partnershipId}"
				onsaved={(count: Count) => void goto(`/app/hitung/${count.id}`)}
			/>
		{/key}
	{/if}
</section>
