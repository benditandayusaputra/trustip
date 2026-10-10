<script lang="ts">
	import Flag from '@lucide/svelte/icons/flag';
	import type { Component } from 'svelte';
	import { listingReport as id } from '#lib/i18n/listing-report.js';
	import type { ReportTarget } from '../report.svelte';

	interface Props {
		target: ReportTarget;
		name: string;
	}

	let { target, name }: Props = $props();

	let Report = $state<Component<{ target: ReportTarget; name: string; startOpen: boolean }>>();

	async function load() {
		Report = (await import('./ReportListing.svelte')).default;
	}
</script>

{#if Report}
	<Report {target} {name} startOpen={true} />
{:else}
	<button
		type="button"
		class="inline-flex min-h-11 items-center gap-2 text-sm text-text-muted"
		onclick={() => void load()}
	>
		<Flag class="size-4" aria-hidden="true" />
		{id.listingReport.open(name)}
	</button>
{/if}
