<script lang="ts">
	import Package from '@lucide/svelte/icons/package';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { DropFilter } from '#lib/api/drops.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { DropHistory, emptyFilter } from '../history.svelte';
	import { statusToneOf } from '../rows';
	import DropFilters from './DropFilters.svelte';

	interface Props {
		partnershipId?: string;
		recordHref?: string | undefined;
	}

	let { partnershipId = '', recordHref }: Props = $props();

	const text = id.drops;
	const history = new DropHistory(apiClient);
	let filter = $derived<DropFilter>(emptyFilter(partnershipId));
	const isFiltered = $derived(filter.status !== '' || filter.from !== '' || filter.to !== '');

	$effect(() => {
		void history.load(filter);
	});
</script>

<DropFilters
	{filter}
	showPartner={partnershipId === ''}
	onchange={(next: DropFilter) => (filter = next)}
/>
{#if history.status === 'loading'}
	<Skeleton />
{:else if history.status === 'failed'}
	<Alert title={text.historyFailed}>{history.failure}</Alert>
{:else if history.items.length === 0}
	<EmptyState icon={Package} message={isFiltered ? text.historyNoMatch : text.historyEmpty}>
		{#snippet action()}
			{#if recordHref}
				<a href={recordHref} class="font-semibold text-accent">{text.recordDrop}</a>
			{/if}
		{/snippet}
	</EmptyState>
{:else}
	<ul class="flex flex-col gap-2">
		{#each history.items as drop (drop.id)}
			<li>
				<a
					href="/app/titipan/{drop.id}"
					class="flex items-center justify-between gap-4 rounded-control border border-border bg-surface p-4 hover:bg-surface-hover"
				>
					<span class="flex flex-col gap-1">
						<span class="font-mono text-sm text-text-muted">{drop.number}</span>
						<span class="text-base font-semibold text-text">{drop.counterpart.name}</span>
						<span class="text-sm text-text-muted">
							{formatDateTime(drop.occurred_at)} · {text.summary(drop.total_qty, drop.line_count)}
						</span>
					</span>
					<Badge tone={statusToneOf(drop.status)} label={text.statuses[drop.status]} />
				</a>
			</li>
		{/each}
	</ul>
	{#if history.nextCursor}
		<Button variant="secondary" busy={history.busy} onclick={() => void history.loadMore()}>
			{text.loadMore}
		</Button>
	{/if}
{/if}
