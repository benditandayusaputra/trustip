<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ConsignorDirectoryView from '#lib/features/directory/components/ConsignorDirectoryView.svelte';
	import {
		consignorAddress,
		consignorQueryFrom,
		type ConsignorQuery
	} from '#lib/features/directory/consignor-query.js';
	import PlacePicker from '#lib/features/organization/components/PlacePicker.svelte';
	import type { Place } from '#lib/features/organization/place.js';
	import { directory as id } from '#lib/i18n/directory.js';

	const query = $derived(consignorQueryFrom(page.url.searchParams));
	const origin = $derived<Place>({
		kind: query.origin,
		lat: query.lat,
		lng: query.lng,
		cityCode: query.cityCode
	});

	function show(next: ConsignorQuery) {
		const address = consignorAddress(next).toString();
		void goto(address === '' ? page.url.pathname : `?${address}`, { replaceState: true });
	}
</script>

<svelte:head>
	<title>{id.consignorDirectory.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.consignorDirectory.title}</h1>
	<ConsignorDirectoryView {query} onchange={show}>
		{#snippet place()}
			<PlacePicker
				place={origin}
				onchange={(chosen: Place) => {
					show({
						...query,
						origin: chosen.kind,
						lat: chosen.lat,
						lng: chosen.lng,
						cityCode: chosen.cityCode
					});
				}}
			/>
		{/snippet}
	</ConsignorDirectoryView>
</section>
