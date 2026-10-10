<script lang="ts">
	import QrCode from '#lib/components/ui/QrCode.svelte';
	import type { Product } from '#lib/api/products.js';
	import { products as id } from '#lib/i18n/products.js';
	import { publicProductUrl } from '../links';

	interface Props {
		product: Product;
	}

	let { product }: Props = $props();

	const text = id.products;
</script>

<section class="flex flex-col gap-2">
	<h2 class="text-lg font-semibold text-text">{text.qrTitle}</h2>
	<p class="text-sm text-text-muted">{text.qrHint}</p>
	<a
		href={publicProductUrl(window.location.origin, product)}
		class="inline-flex min-h-11 items-center self-start font-semibold text-accent"
	>
		{text.openPublicPage}
	</a>
	<QrCode
		value={publicProductUrl(window.location.origin, product)}
		alt={text.qrAlt(product.name)}
		fileName="{product.sku}-qr.svg"
		downloadLabel={text.downloadQr}
	/>
</section>
