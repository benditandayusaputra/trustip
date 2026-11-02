<script lang="ts">
	import ContentSkeleton from '../ui/ContentSkeleton.svelte';
	import DashboardSkeleton from '../ui/DashboardSkeleton.svelte';
	import Skeleton from '../ui/Skeleton.svelte';
	import { pageSkeletonKind } from './page-skeleton';
	let { path, titled = true }: { path: string; titled?: boolean } = $props();
	const kind = $derived(pageSkeletonKind(path));
	const dashboard = $derived(kind === 'store-dashboard' || kind === 'consignor-dashboard');
	const narrow = $derived(['form', 'auth'].includes(kind));
	const reading = $derived(kind === 'article');
</script>

<div
	class="mx-auto flex w-full min-w-0 flex-col gap-6 {narrow ? 'max-w-form' : ''} {reading
		? 'max-w-reading'
		: ''}"
>
	{#if titled || dashboard}
		<div aria-hidden="true" class="flex flex-col gap-4">
			<div class="w-1/2"><Skeleton shape="metric" /></div>
			<div class="w-2/3"><Skeleton /></div>
		</div>
	{/if}
	{#if kind === 'store-dashboard' || kind === 'consignor-dashboard'}
		<DashboardSkeleton side={kind === 'store-dashboard' ? 'STORE' : 'CONSIGNOR'} />
	{:else}
		<ContentSkeleton variant={kind} />
	{/if}
</div>
