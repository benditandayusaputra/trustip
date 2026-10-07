<script lang="ts">
	import Download from '@lucide/svelte/icons/download';
	import Skeleton from './Skeleton.svelte';

	interface Props {
		value: string;
		alt: string;
		fileName: string;
		downloadLabel: string;
	}

	let { value, alt, fileName, downloadLabel }: Props = $props();

	const displaySize = 192;
	let source = $state('');

	$effect(() => {
		const encoded = value;
		void import('uqr').then(({ renderSVG }) => {
			source = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(renderSVG(encoded))}`;
		});
	});
</script>

{#if source === ''}
	<Skeleton />
{:else}
	<div class="flex flex-col items-start gap-2">
		<img src={source} {alt} width={displaySize} height={displaySize} class="rounded-control" />
		<a
			href={source}
			download={fileName}
			class="inline-flex min-h-12 items-center gap-2 rounded-control border border-border-strong px-5 font-semibold text-text"
		>
			<Download class="size-5" aria-hidden="true" />
			{downloadLabel}
		</a>
	</div>
{/if}
