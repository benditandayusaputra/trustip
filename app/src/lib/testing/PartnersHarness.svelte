<script lang="ts">
	import PartnerList from '#lib/features/partnerships/components/PartnerList.svelte';
	import PartnerPage from '#lib/features/partnerships/components/PartnerPage.svelte';

	interface Props {
		view: 'list' | 'page';
		partnershipId?: string;
	}

	let { view, partnershipId = '' }: Props = $props();
</script>

{#if view === 'list'}
	<PartnerList />
{:else}
	<PartnerPage {partnershipId} canAmend={true} isConsignor={true}>
		{#snippet agreement()}
			<p>Kesepakatan tampil di sini</p>
		{/snippet}
	</PartnerPage>
{/if}
