<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Search from '@lucide/svelte/icons/search';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import { positionOf } from '#lib/features/directory/visitor.js';
	import { locateDevice } from '#lib/features/organization/geolocation.js';
	import ConsignorDirectoryView from '#lib/features/directory/components/ConsignorDirectoryView.svelte';
	import {
		consignorAddress,
		consignorQueryFrom,
		type ConsignorQuery
	} from '#lib/features/directory/consignor-query.js';
	import PlacePicker from '#lib/features/organization/components/PlacePicker.svelte';
	import type { Place } from '#lib/features/organization/place.js';
	import { directory as id } from '#lib/i18n/directory.js';

	const session = getAccountSession();
	const query = $derived(consignorQueryFrom(page.url.searchParams));
	const origin = $derived<Place>({
		kind: query.origin,
		lat: query.lat,
		lng: query.lng,
		cityCode: query.cityCode
	});

	const isVisitor = $derived(session.status === 'signedOut');
	let askedPosition = false;

	$effect(() => {
		if (!isVisitor || askedPosition || query.origin !== 'ORGANIZATION') {
			return;
		}
		askedPosition = true;
		locateDevice()
			.then((point) => {
				show({ ...query, ...positionOf(point) });
			})
			.catch(() => undefined);
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
	<PageHeading icon={Search}>{id.consignorDirectory.title}</PageHeading>
	<ConsignorDirectoryView
		{query}
		{isVisitor}
		ready={session.status !== 'loading' || session.expectedWorkspace !== undefined}
		onchange={show}
	>
		{#snippet place()}
			<PlacePicker
				place={origin}
				{isVisitor}
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
