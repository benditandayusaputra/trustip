<script lang="ts">
	import { formatDateTime } from '#lib/i18n/format.js';
	import { offline as id } from '#lib/i18n/offline.js';
	import Badge from './Badge.svelte';

	interface Props {
		recordedAt: string;
		receivedAt: string;
	}

	let { recordedAt, receivedAt }: Props = $props();
</script>

<div class="flex flex-wrap items-center gap-2 text-sm text-text-muted">
	<Badge tone="pending" label={id.offline.recordedOffline} />
	<span>{id.offline.bothTimes(formatDateTime(recordedAt), formatDateTime(receivedAt))}</span>
</div>
