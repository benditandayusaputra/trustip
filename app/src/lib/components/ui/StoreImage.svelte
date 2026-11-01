<script lang="ts">
	let { url, class: className = '' }: { url: string | null; class?: string } = $props();

	let failedUrl = $state<string | null>(null);
	const showsPhoto = $derived(url !== null && url !== failedUrl);
</script>

{#if showsPhoto}
	<img
		src={url}
		alt=""
		width="48"
		height="48"
		loading="lazy"
		decoding="async"
		onerror={() => (failedUrl = url)}
		class="shrink-0 object-cover {className}"
	/>
{:else}
	<svg
		viewBox="0 0 48 48"
		aria-hidden="true"
		data-default-store-image
		class="shrink-0 overflow-hidden {className}"
	>
		<rect width="48" height="48" fill="var(--color-accent-surface)" />
		<ellipse cx="24" cy="39" rx="18" ry="2" fill="var(--color-illustration-ground)" />
		<rect
			x="12"
			y="17"
			width="24"
			height="22"
			rx="1.5"
			fill="var(--color-surface)"
			stroke="var(--color-border)"
			stroke-width="0.75"
		/>
		<path d="M10 13l3-6h22l3 6Z" fill="var(--color-primary)" />
		<path d="M18 7h4l-1 6h-4Zm8 0h4l2 6h-4Z" fill="var(--color-primary-hover)" />
		<path
			d="M10 13h28v2a2.8 2.8 0 0 1-5.6 0 2.8 2.8 0 0 1-5.6 0 2.8 2.8 0 0 1-5.6 0 2.8 2.8 0 0 1-5.6 0 2.8 2.8 0 0 1-5.6 0Z"
			fill="var(--color-primary-hover)"
		/>
		<rect x="16" y="24" width="7" height="15" rx="1" fill="var(--color-accent-surface)" />
		<rect x="26" y="23" width="7" height="6" rx="1" fill="var(--color-accent-surface)" />
		<circle cx="21.5" cy="32" r="0.8" fill="var(--color-primary)" />
	</svg>
{/if}
