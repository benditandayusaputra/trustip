<script lang="ts">
	import Close from '@lucide/svelte/icons/x';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import { organization as id } from '#lib/i18n/organization.js';
	import { maxStorePhotos, type StoreProfileValues } from '../store-profile.svelte';

	interface Props {
		photos: StoreProfileValues['photos'];
		busy: boolean;
		error: string | undefined;
		onadd: (file: File) => void;
		onremove: (photoId: string) => void;
	}

	let { photos, busy, error, onadd, onremove }: Props = $props();

	const text = id.storeProfile;
	const thumbnailSize = 48;
</script>

<div class="flex flex-col gap-2">
	{#if photos.length > 0}
		<ul class="flex flex-wrap gap-2">
			{#each photos as photo, index (photo.id)}
				<li class="relative">
					<img
						src={photo.url}
						alt={text.photoAlt(index + 1)}
						width={thumbnailSize}
						height={thumbnailSize}
						loading="lazy"
						decoding="async"
						class="size-12 rounded-control border border-border object-cover"
					/>
					<button
						type="button"
						aria-label={text.removePhoto(index + 1)}
						onclick={() => {
							onremove(photo.id);
						}}
						class="inline-flex size-11 items-center justify-center rounded-pill text-text"
					>
						<Close class="size-5" aria-hidden="true" />
					</button>
				</li>
			{/each}
		</ul>
	{/if}
	{#if photos.length < maxStorePhotos}
		<ImageField
			label={text.photos}
			chooseLabel={text.addPhoto}
			imageUrl={undefined}
			imageAlt=""
			hint={text.photosHint}
			{error}
			{busy}
			onselect={onadd}
		/>
	{/if}
</div>
