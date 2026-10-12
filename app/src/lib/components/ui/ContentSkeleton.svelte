<script lang="ts">
	import { common as id } from '#lib/i18n/common.js';
	import Skeleton from './Skeleton.svelte';
	import SkeletonRows from './SkeletonRows.svelte';
	import SkeletonFields from './SkeletonFields.svelte';
	import SkeletonCards from './SkeletonCards.svelte';
	import SkeletonReport from './SkeletonReport.svelte';
	import type { SkeletonVariant } from './skeleton';
	interface Props {
		variant?: SkeletonVariant;
		rows?: number;
	}
	let { variant = 'list', rows = 4 }: Props = $props();
	const panels = 2;
	const paragraphLines = 3;
</script>

<div
	role="status"
	aria-label={id.feedback.loading}
	aria-busy="true"
	data-loading-skeleton={variant}
	class="w-full min-w-0"
>
	<div aria-hidden="true" class="flex min-w-0 flex-col gap-6">
		{#if variant === 'list'}
			<SkeletonRows {rows} />
		{:else if variant === 'cards'}
			<SkeletonCards {rows} />
		{:else if variant === 'form' || variant === 'auth'}
			<SkeletonFields {rows} />
		{:else if variant === 'report'}
			<SkeletonReport {rows} />
		{:else if variant === 'product' || variant === 'home'}
			<div class="grid items-center gap-6 lg:grid-cols-2">
				<Skeleton shape="image" />
				<div class="flex flex-col gap-6">
					<div class="w-1/2"><Skeleton /></div>
					<Skeleton shape="control" /><Skeleton lines={paragraphLines} />
					<div class="w-1/2"><Skeleton shape="control" /></div>
				</div>
			</div>
			<SkeletonCards rows={panels} />
		{:else if variant === 'article'}
			<div class="flex flex-col gap-8">
				{#each { length: rows }, index (index)}
					<div class="flex flex-col gap-4">
						<div class="w-1/3"><Skeleton shape="metric" /></div>
						<Skeleton lines={paragraphLines} />
						<div class="w-2/3"><Skeleton /></div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="flex flex-col gap-6 rounded-card border border-border bg-surface p-5">
				{#if variant === 'profile'}
					<div class="flex items-center gap-4">
						<div class="size-12 shrink-0"><Skeleton shape="square" /></div>
						<div class="w-2/3"><Skeleton shape="metric" /></div>
					</div>
				{:else}
					<div class="w-1/2"><Skeleton shape="metric" /></div>
				{/if}
				<Skeleton lines={rows} />
				<div class="grid grid-cols-2 gap-4"><Skeleton /><Skeleton /></div>
				<div class="w-1/2"><Skeleton shape="control" /></div>
			</div>
			<SkeletonRows rows={panels} />
		{/if}
	</div>
</div>
