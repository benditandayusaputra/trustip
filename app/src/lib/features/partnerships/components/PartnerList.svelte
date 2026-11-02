<script lang="ts">
	import Handshake from '@lucide/svelte/icons/handshake';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import Tabs from '#lib/components/ui/Tabs.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDate, formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { partnerships as id } from '#lib/i18n/partnerships.js';
	import { PartnerList, type PartnerFilter } from '../state.svelte';

	const text = id.partners;
	const list = new PartnerList(apiClient);
	const tabs = Object.entries(text.filters).map(([key, label]) => ({ key, label }));
	let selected = $state<string>('ACTIVE');

	$effect(() => {
		void list.load(selected as PartnerFilter);
	});
</script>

<Tabs label={text.filterLabel} {tabs} bind:selected param="tab">
	{#snippet panel()}
		{#if list.status === 'loading'}
			<ContentSkeleton variant="list" />
		{:else if list.status === 'failed'}
			<Alert title={text.loadFailed}>{list.failure}</Alert>
		{:else if list.items.length === 0}
			<EmptyState icon={Handshake} message={text.empty}>
				{#snippet action()}
					<a href="/app/ajuan/baru" class="font-semibold text-accent">{text.invite}</a>
				{/snippet}
			</EmptyState>
		{:else}
			<ul class="flex flex-col gap-2">
				{#each list.items as partner (partner.id)}
					<li>
						<a
							href="/app/mitra/{partner.id}"
							class="surface-card flex items-center justify-between gap-4 p-5 hover:border-accent hover:bg-surface-hover"
						>
							<span class="flex flex-col gap-1">
								<span class="text-base font-semibold text-text">{partner.counterpart.name}</span>
								<span class="text-sm text-text-muted">
									{text.types[partner.counterpart.type]} · {text.since(
										formatDate(partner.started_at)
									)}
								</span>
								<span class="font-mono text-sm text-text tabular-nums">
									{text.balance(
										formatInteger(partner.stock_qty),
										formatRupiah(partner.unpaid_amount)
									)}
								</span>
							</span>
							<Badge
								tone={partner.status === 'ACTIVE' ? 'success' : 'neutral'}
								label={text.statuses[partner.status]}
							/>
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
	{/snippet}
</Tabs>
