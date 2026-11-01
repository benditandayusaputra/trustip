<script lang="ts">
	import type { Snippet } from 'svelte';
	import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
	import PublicMenu from './PublicMenu.svelte';
	import Brand from './Brand.svelte';
	import PublicLinks from './PublicLinks.svelte';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as id } from '#lib/i18n/common.js';

	let {
		currentPath,
		actions
	}: { currentPath?: string | undefined; actions?: Snippet | undefined } = $props();
</script>

<header class="public-header">
	<div
		class="mx-auto flex w-full max-w-content items-center justify-between gap-4 px-4 py-2 lg:px-8"
	>
		<Brand />
		<nav aria-label={id.layout.publicNavigation} class="flex items-center gap-2">
			<PublicLinks {currentPath} />
		</nav>
		<div class="flex items-center gap-2">
			<div class="hidden md:block"><ThemePicker /></div>
			{#if actions}
				{@render actions()}
			{:else}
				<a
					href="/masuk"
					class="inline-flex min-h-11 items-center gap-2 rounded-control bg-primary px-4 text-sm font-semibold text-on-accent hover:bg-primary-hover"
				>
					{id.nav.signIn}<ArrowUpRight class="size-4" aria-hidden="true" />
				</a>
			{/if}
			<PublicMenu />
		</div>
	</div>
</header>
