<script lang="ts">
	import { onMount } from 'svelte';
	import WifiOff from '@lucide/svelte/icons/wifi-off';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { offline as id } from '#lib/i18n/offline.js';
	import { watchOnline } from '#lib/offline/network.js';
	import type { Freshness } from '#lib/offline/freshness.svelte.js';
	import { deviceReadCache } from '#lib/offline/read-cache.js';

	interface Props {
		freshness?: Freshness;
	}

	let { freshness = deviceReadCache().freshness }: Props = $props();

	const text = id.offline;
	let isOnline = $state(true);

	onMount(() => watchOnline((online) => (isOnline = online)));
</script>

{#if !isOnline || freshness.staleSince !== null}
	<div
		role="status"
		class="flex items-center gap-2 border-b border-border bg-warning-surface px-4 py-2 text-sm text-text"
	>
		<WifiOff class="size-4 shrink-0 text-warning" aria-hidden="true" />
		<p>
			{isOnline ? text.staleRibbon : text.ribbon}
			{#if freshness.staleSince !== null}
				{text.storedAt(formatDateTime(freshness.staleSince))}
			{/if}
		</p>
	</div>
{/if}
