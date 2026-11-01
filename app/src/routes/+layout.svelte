<script lang="ts">
	import { onMount } from 'svelte';
	import '#lib/styles/app.css';
	import interRegular from '#lib/assets/fonts/inter-latin-400-normal.woff2';
	import { themeColor } from '#lib/styles/tokens.js';
	import UpdateNotice from '#lib/features/system/UpdateNotice.svelte';
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import { setSessionSlot } from '#lib/features/account/session-slot.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
	let pageThemeColor = $state(themeColor);

	setToaster(new ToastQueue());
	setSessionSlot({});

	function readThemeColor() {
		pageThemeColor = getComputedStyle(document.documentElement)
			.getPropertyValue('--color-bg')
			.trim();
	}

	onMount(() => {
		readThemeColor();
		window.addEventListener('trustip:appearance', readThemeColor);
		return () => {
			window.removeEventListener('trustip:appearance', readThemeColor);
		};
	});
</script>

<svelte:head>
	<link rel="preload" href={interRegular} as="font" type="font/woff2" crossorigin="anonymous" />
	<meta name="theme-color" content={pageThemeColor} />
	<link rel="manifest" href="/manifest.webmanifest" />
	<link rel="icon" href="/icons/favicon.png" type="image/png" />
	<link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />
</svelte:head>

<UpdateNotice />
{@render children()}
<Toaster />
