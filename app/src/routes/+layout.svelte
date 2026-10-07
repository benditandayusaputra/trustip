<script lang="ts">
	import '#lib/styles/app.css';
	import interRegular from '#lib/assets/fonts/inter-latin-400-normal.woff2';
	import { themeColor } from '#lib/styles/tokens.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<svelte:head>
	<link rel="preload" href={interRegular} as="font" type="font/woff2" crossorigin="anonymous" />
	<meta name="theme-color" content={themeColor} />
</svelte:head>

{@render children()}
