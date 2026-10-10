<script lang="ts">
	import SealVerificationView from '#lib/features/seals/components/SealVerificationView.svelte';
	import { agreements as id } from '#lib/i18n/agreements.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>{id.seal.verifyTitle} {data.code}</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<SealVerificationView code={data.code} verification={data.verification} />
