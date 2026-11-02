<script lang="ts">
	import MapPin from '@lucide/svelte/icons/map-pin';
	import Store from '@lucide/svelte/icons/store';

	interface Props {
		kind: string;
		place: string;
		logoUrl: string | null;
	}

	let { kind, place, logoUrl }: Props = $props();
</script>

<header class="trade-panel flex flex-wrap items-center gap-5">
	{#if logoUrl}
		<img
			src={logoUrl}
			alt=""
			width="48"
			height="48"
			decoding="async"
			class="size-12 shrink-0 rounded-control border border-border object-cover"
		/>
	{:else}
		<span
			class="flex size-12 shrink-0 items-center justify-center rounded-control bg-accent-surface"
			><Store class="size-6 text-accent" aria-hidden="true" /></span
		>
	{/if}
	<div class="flex min-w-0 flex-col gap-1">
		<p class="section-label">{kind}</p>
		{#if place}
			<p class="flex items-center gap-1 text-sm text-text-muted">
				<MapPin class="size-4 shrink-0" aria-hidden="true" />{place}
			</p>
		{/if}
	</div>
</header>
