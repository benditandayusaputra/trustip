<script lang="ts">
	import Store from '@lucide/svelte/icons/store';

	interface Props {
		name: string;
		kind: string;
		place: string;
		logoUrl: string | null;
	}

	let { name, kind, place, logoUrl }: Props = $props();
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
		<Store class="size-12 shrink-0 text-text-muted" aria-hidden="true" />
	{/if}
	<div class="flex min-w-0 flex-col gap-1">
		<p class="section-label">{kind}</p>
		<h1 class="text-xl font-semibold text-text">{name}</h1>
		{#if place}
			<p class="text-sm text-text-muted">{place}</p>
		{/if}
	</div>
</header>
