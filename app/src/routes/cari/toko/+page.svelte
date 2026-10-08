<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import StoreDirectoryView from '#lib/features/directory/components/StoreDirectoryView.svelte';
	import {
		storeQueryFrom,
		storeSearchParams,
		type StoreQuery
	} from '#lib/features/directory/query.js';
	import PlacePicker from '#lib/features/organization/components/PlacePicker.svelte';
	import type { Place } from '#lib/features/organization/place.js';
	import { id } from '#lib/i18n/id.js';

	const query = $derived(storeQueryFrom(page.url.searchParams));
	const origin = $derived<Place>({
		kind: query.origin,
		lat: query.lat,
		lng: query.lng,
		cityCode: query.cityCode
	});

	function show(next: StoreQuery) {
		const search = storeSearchParams(next).toString();
		void goto(search === '' ? page.url.pathname : `?${search}`, { replaceState: true });
	}
</script>

<svelte:head>
	<title>{id.directory.storesTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.directory.storesTitle}</h1>
	<StoreDirectoryView {query} onchange={show}>
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
	</StoreDirectoryView>
</section>
