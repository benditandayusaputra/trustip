<script lang="ts">
	import { common as id } from '#lib/i18n/common.js';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	interface Props {
		side?: 'STORE' | 'CONSIGNOR';
	}
	let { side = 'STORE' }: Props = $props();
	const storeCards = 5;
	const consignorCards = 4;
	const rowCount = 3;
	const cardCount = $derived(side === 'STORE' ? storeCards : consignorCards);
</script>

<div
	role="status"
	aria-label={id.feedback.loading}
	aria-busy="true"
	data-loading-skeleton="dashboard"
	class="flex w-full flex-col gap-6"
>
	<div
		aria-hidden="true"
		class="grid grid-cols-2 gap-4 {side === 'STORE'
			? 'md:grid-cols-3 xl:grid-cols-5'
			: 'lg:grid-cols-4'}"
	>
		{#each { length: cardCount }, index (index)}
			<div
				class="flex min-w-0 flex-col gap-4 rounded-card border border-border bg-surface p-5 {index >
				1
					? 'col-span-full compact:col-span-1'
					: ''}"
			>
				<div class="flex items-center justify-between gap-4">
					<div class="w-1/2"><Skeleton /></div>
					<div class="size-11"><Skeleton shape="control" /></div>
				</div>
				<Skeleton shape="control" /><Skeleton />
			</div>
		{/each}
	</div>
	<div aria-hidden="true" class="grid gap-6 lg:grid-cols-12">
		<div class="flex flex-col gap-6 rounded-card border border-border bg-surface p-5 lg:col-span-8">
			<div class="w-1/2"><Skeleton /></div>
			<Skeleton lines={consignorCards} shape="control" />
		</div>
		<div class="flex flex-col gap-6 rounded-card border border-border bg-surface p-5 lg:col-span-4">
			<div class="w-1/2"><Skeleton /></div>
			<Skeleton lines={rowCount} shape="control" />
		</div>
	</div>
	<div
		aria-hidden="true"
		class="flex flex-col gap-6 rounded-card border border-border bg-surface p-5"
	>
		<div class="w-1/4"><Skeleton /></div>
		<div class="grid grid-cols-(--insight-columns) gap-4">
			{#each { length: rowCount }, index (index)}
				<div class="flex flex-col gap-4 rounded-control border border-border p-4">
					<div class="w-1/2"><Skeleton shape="control" /></div>
					<Skeleton lines={rowCount} /><Skeleton shape="control" />
				</div>
			{/each}
		</div>
	</div>
	{#each { length: 2 }, index (index)}
		<div
			aria-hidden="true"
			class="flex flex-col gap-6 rounded-card border border-border bg-surface p-5"
		>
			<div class="w-1/4"><Skeleton /></div>
			<Skeleton lines={rowCount} shape="control" />
		</div>
	{/each}
</div>
