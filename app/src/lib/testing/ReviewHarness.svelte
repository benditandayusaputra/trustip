<script lang="ts">
	import OutboxReview from '#lib/features/outbox/components/OutboxReview.svelte';

	interface Props {
		entryId: string;
		ondiscarded?: () => void;
	}

	let { entryId, ondiscarded = () => undefined }: Props = $props();
</script>

<OutboxReview {entryId} {ondiscarded} />
