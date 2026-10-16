<script lang="ts">
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Product } from '#lib/api/products.js';
	import { products as id } from '#lib/i18n/products.js';
	import { ProductRecord } from '../catalog.svelte';
	import ProductForm from './ProductForm.svelte';
	import ProductQr from './ProductQr.svelte';
	import Switch from '#lib/components/ui/Switch.svelte';

	interface Props {
		productId: string;
	}

	let { productId }: Props = $props();

	const text = id.products;
	const record = new ProductRecord(apiClient);
	const toaster = getToaster();

	$effect(() => {
		void record.load(productId);
	});

	async function setSeeking(isSeeking: boolean) {
		try {
			await record.setSeeking(isSeeking);
			toaster.show({ tone: 'success', message: isSeeking ? text.seekingOn : text.seekingOff });
		} catch (error) {
			toaster.show({ tone: 'danger', message: error instanceof Error ? error.message : '' });
		}
	}

	async function toggleArchive() {
		try {
			const changed = await record.toggleArchive();
			toaster.show({
				tone: 'success',
				message: changed.status === 'ARCHIVED' ? text.archived : text.reactivated
			});
		} catch (error) {
			toaster.show({ tone: 'danger', message: error instanceof Error ? error.message : '' });
		}
	}
</script>

{#if record.status === 'loading'}
	<ContentSkeleton variant="form" />
{:else if record.status === 'failed' || !record.product}
	<Alert title={text.loadFailed}>
		{record.failure}
		{#snippet action()}
			<a href="/app/produk" class="font-semibold text-accent">{text.backToCatalog}</a>
		{/snippet}
	</Alert>
{:else}
	{@const product = record.product}
	<div class="flex flex-wrap items-center justify-between gap-4">
		<p class="flex items-center gap-2 text-sm text-text-muted">
			{product.sku}
			<Badge
				tone={product.status === 'ACTIVE' ? 'success' : 'neutral'}
				label={text.statuses[product.status]}
			/>
		</p>
		<Button
			variant={product.status === 'ACTIVE' ? 'danger' : 'secondary'}
			busy={record.busy}
			onclick={() => void toggleArchive()}
		>
			{product.status === 'ACTIVE' ? text.archive : text.reactivate}
		</Button>
	</div>
	{#if product.status === 'ARCHIVED'}
		<p role="status" class="text-sm text-warning">{text.archivedNotice}</p>
	{/if}
	<Switch
		label={text.seeking}
		description={text.seekingHint}
		checked={product.is_seeking_stores}
		onchange={(isSeeking: boolean) => void setSeeking(isSeeking)}
	/>
	{#key product.id}
		<ProductForm
			{product}
			onsaved={(saved: Product) => {
				record.product = saved;
			}}
		/>
	{/key}
	<details class="secondary-disclosure">
		<summary
			>{text.qrTitle}<ChevronDown class="size-5 text-text-muted" aria-hidden="true" /></summary
		>
		<ProductQr {product} />
	</details>
{/if}
