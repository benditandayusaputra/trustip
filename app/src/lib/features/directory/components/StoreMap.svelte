<script lang="ts">
	import MapPin from '@lucide/svelte/icons/map-pin';
	import { type Component, type ComponentProps } from 'svelte';
	import type MapView from '#lib/components/ui/MapView.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import StoreCard from './StoreCard.svelte';
	type LocatedStore = DirectoryStoreCard & {
		location: NonNullable<DirectoryStoreCard['location']>;
	};
	interface Props {
		stores: DirectoryStoreCard[];
	}
	let { stores }: Props = $props();
	const text = id.directory;
	let Map = $state<Component<ComponentProps<typeof MapView>>>();
	let failed = $state(false);
	let selectedId = $state<string>();
	const located = $derived(
		stores.filter((store): store is LocatedStore => store.location !== null)
	);
	const markers = $derived(
		located.map((store) => ({
			id: store.id,
			label: text.markerLabel(store.name),
			...store.location
		}))
	);
	const center = $derived(located[0]?.location);
	const selected = $derived(located.find((store) => store.id === selectedId));
	$effect(() => {
		if (!center || Map || failed) return;
		let mounted = true;
		import('#lib/components/ui/MapView.svelte')
			.then((module) => {
				if (mounted) Map = module.default;
			})
			.catch(() => {
				if (mounted) failed = true;
			});
		return () => {
			mounted = false;
		};
	});
</script>

{#if !center}
	<p role="status" class="rounded-card border border-border bg-surface p-4 text-text-muted">
		{text.noMapLocations}
	</p>
	<ul class="grid grid-cols-(--store-results-columns) items-start gap-4">
		{#each stores as store (store.id)}<li><StoreCard card={store} /></li>{/each}
	</ul>
{:else}
	<div class="flex min-w-0 flex-col gap-4">
		<p class="text-sm text-text-muted">{text.mapHint}</p>
		<div class="grid items-start gap-4 xl:grid-cols-2">
			<div class="flex min-w-0 flex-col gap-4">
				{#if Map}<Map
						{center}
						label={text.mapLabel}
						unavailableText={text.mapUnavailable}
						groupLabel={text.markerGroup}
						onmove={() => undefined}
						{markers}
						{selectedId}
						onselect={(id: string) => (selectedId = id)}
					/>
				{:else if failed}<p role="status" class="text-warning">{text.mapUnavailable}</p>
				{:else}<Skeleton />{/if}
				<p class="text-sm text-text-muted">{text.mapCount(located.length)}</p>
				{#if located.length < stores.length}<p class="text-sm text-text-muted">
						{text.missingLocations(stores.length - located.length)}
					</p>{/if}
				<details class="rounded-control border border-border bg-surface p-4">
					<summary class="min-h-11 content-center font-semibold text-accent"
						>{text.storeChoices} ({located.length})</summary
					>
					<ul
						aria-label={text.storeChoices}
						class="mt-4 grid max-h-(--directory-map-height) gap-2 overflow-y-auto compact:grid-cols-2"
					>
						{#each located as store (store.id)}
							<li>
								<button
									type="button"
									aria-pressed={store.id === selectedId}
									onclick={() => (selectedId = store.id)}
									class="flex min-h-12 w-full items-center gap-2 rounded-control border px-4 py-2 text-left text-sm {store.id ===
									selectedId
										? 'border-accent bg-accent-surface text-accent'
										: 'border-border bg-surface text-text hover:bg-surface-hover'}"
								>
									<MapPin class="size-5 shrink-0" aria-hidden="true" /><span
										class="min-w-0 break-words">{store.name}</span
									>
								</button>
							</li>
						{/each}
					</ul>
				</details>
			</div>
			<section aria-label={text.selectedStore} class="min-w-0">
				{#if selected}<StoreCard card={selected} />
				{:else}<p
						class="rounded-card border border-dashed border-border-strong bg-surface p-5 text-text-muted"
					>
						{text.chooseMarker}
					</p>{/if}
			</section>
		</div>
		<p role="status" class="sr-only">{selected ? text.selectedAnnouncement(selected.name) : ''}</p>
	</div>
{/if}
