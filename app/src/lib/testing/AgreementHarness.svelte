<script lang="ts">
	import AgreementView from '#lib/features/agreements/components/AgreementView.svelte';

	interface Props {
		partnershipId: string;
	}

	let { partnershipId }: Props = $props();
</script>

<AgreementView {partnershipId} />
