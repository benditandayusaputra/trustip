<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Drop } from '#lib/api/drops.js';
	import DropForm from '#lib/features/drops/components/DropForm.svelte';
	import PartnerChooser from '#lib/features/drops/components/PartnerChooser.svelte';
	import { id } from '#lib/i18n/id.js';

	const partnershipId = $derived(page.url.searchParams.get('mitra') ?? '');
</script>

<svelte:head>
	<title>{id.drops.newTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.drops.newTitle}</h1>
	<p class="text-base text-text-muted">{id.drops.newIntro}</p>
	{#if partnershipId === ''}
		<PartnerChooser hrefOf={(partnerId: string) => `/app/titipan/baru?mitra=${partnerId}`} />
	{:else}
		{#key partnershipId}
			<DropForm {partnershipId} onsaved={(drop: Drop) => void goto(`/app/titipan/${drop.id}`)} />
		{/key}
	{/if}
</section>
