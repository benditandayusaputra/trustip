<script lang="ts">
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { NeedStatus } from '#lib/api/needs.js';
	import { formatDate } from '#lib/i18n/format.js';
	import { needs as id } from '#lib/i18n/needs.js';
	import { NeedList } from '../state.svelte';
	import { needTone } from '../tone';

	const text = id.needs;
	const list = new NeedList(apiClient);
	let filter = $state<NeedStatus | ''>('');
	const statusOptions = [
		{ value: '', label: text.allStatuses },
		...Object.entries(text.statuses).map(([value, label]) => ({ value, label }))
	];

	$effect(() => {
		void list.load(filter);
	});
</script>

<Select
	label={text.status}
	options={statusOptions}
	value={filter}
	onchange={(status: string) => (filter = status as NeedStatus | '')}
/>
{#if list.status === 'loading'}
	<ContentSkeleton variant="list" />
{:else if list.status === 'failed'}
	<Alert title={text.failed}>{list.failure}</Alert>
{:else if list.items.length === 0}
	<EmptyState icon={Megaphone} message={filter === '' ? text.empty : text.noMatch}>
		{#snippet action()}
			<a href="/app/kebutuhan/baru" class="font-semibold text-accent">{text.post}</a>
		{/snippet}
	</EmptyState>
{:else}
	<ul class="flex flex-col gap-2">
		{#each list.items as need (need.id)}
			<li>
				<a
					href="/app/kebutuhan/{need.id}"
					class="flex items-center justify-between gap-4 rounded-control border border-border bg-surface p-4 hover:bg-surface-hover"
				>
					<span class="flex flex-col gap-1">
						<span class="text-base font-semibold text-text">{need.title}</span>
						<span class="text-sm text-text-muted">
							{need.category.name} · {text.expiresAt(formatDate(need.expires_at))} ·
							{text.responses(need.response_count)}
						</span>
					</span>
					<Badge tone={needTone(need.status)} label={text.statuses[need.status]} />
				</a>
			</li>
		{/each}
	</ul>
	{#if list.nextCursor}
		<Button variant="secondary" busy={list.busy} onclick={() => void list.loadMore()}>
			{text.loadMore}
		</Button>
	{/if}
{/if}
