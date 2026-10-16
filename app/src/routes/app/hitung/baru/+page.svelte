<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Calculator from '@lucide/svelte/icons/calculator';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import CountForm from '#lib/features/counts/components/CountForm.svelte';
	import { expiredReturnOf } from '#lib/features/counts/schema.js';
	import PartnerChooser from '#lib/features/drops/components/PartnerChooser.svelte';
	import { counts as id } from '#lib/i18n/counts.js';

	const partnershipId = $derived(page.url.searchParams.get('mitra') ?? '');
	const revisionOf = $derived(page.url.searchParams.get('antrean') ?? '');
	const expiredReturn = $derived(expiredReturnOf(page.url.searchParams));
</script>

<svelte:head>
	<title>{id.counts.newTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<PageHeading backHref="/app/hitung" icon={Calculator}>{id.counts.newTitle}</PageHeading>
	<p class="text-base text-text-muted">{id.counts.newIntro}</p>
	{#if partnershipId === ''}
		<PartnerChooser hrefOf={(partnerId: string) => `/app/hitung/baru?mitra=${partnerId}`} />
	{:else}
		{#key partnershipId}
			<CountForm
				{partnershipId}
				{revisionOf}
				{expiredReturn}
				dropHref="/app/titipan/baru?mitra={partnershipId}"
				onsaved={(countId: string) => void goto(`/app/hitung/${countId}`)}
				onqueued={() => void goto('/app/antrean')}
			/>
		{/key}
	{/if}
</section>
