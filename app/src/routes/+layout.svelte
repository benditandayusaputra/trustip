<script lang="ts">
	import '#lib/styles/app.css';
	import interRegular from '#lib/assets/fonts/inter-latin-400-normal.woff2';
	import { themeColor } from '#lib/styles/tokens.js';
	import UpdateNotice from '#lib/features/system/UpdateNotice.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="preload" href={interRegular} as="font" type="font/woff2" crossorigin="anonymous" />
	<meta name="theme-color" content={themeColor} />
	<link rel="manifest" href="/manifest.webmanifest" />
	<link rel="icon" href="/icons/icon.svg" type="image/svg+xml" />
	<link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />
</svelte:head>

<UpdateNotice />
{@render children()}
