<script lang="ts">
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import Handshake from '@lucide/svelte/icons/handshake';
	import Alert from '#lib/components/ui/Alert.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { drops as dropsText } from '#lib/i18n/drops.js';
	import { PartnerOptions } from '../partners.svelte';

	const id = { ...commonText, ...dropsText };

	interface Props {
		hrefOf: (partnershipId: string) => string;
	}

	let { hrefOf }: Props = $props();

	const text = id.drops;
	const partners = new PartnerOptions(apiClient);

	$effect(() => {
		void partners.load();
	});
</script>

{#if partners.status === 'loading'}
	<Skeleton />
{:else if partners.status === 'failed'}
	<Alert title={text.partnersFailed}>{partners.failure}</Alert>
{:else if partners.items.length === 0}
	<EmptyState icon={Handshake} message={text.noPartners}>
		{#snippet action()}
			<a href="/app/ajuan" class="font-semibold text-accent">{id.nav.proposals}</a>
		{/snippet}
	</EmptyState>
{:else}
	<nav aria-label={text.choosePartner} class="flex flex-col gap-2">
		<h2 class="text-lg font-semibold text-text">{text.choosePartner}</h2>
		<p class="text-sm text-text-muted">{text.choosePartnerHint}</p>
		<ul class="flex flex-col gap-2">
			{#each partners.items as partner (partner.id)}
				<li>
					<a
						href={hrefOf(partner.id)}
						class="flex min-h-12 items-center justify-between rounded-card border border-border bg-surface px-4 py-2 text-base font-semibold text-text hover:bg-surface-hover"
					>
						{partner.counterpart.name}
						<ChevronRight class="size-5 text-text-muted" aria-hidden="true" />
					</a>
				</li>
			{/each}
		</ul>
	</nav>
{/if}
