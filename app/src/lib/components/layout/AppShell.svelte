<script lang="ts">
	import type { Snippet } from 'svelte';
	import BottomNav from './BottomNav.svelte';
	import Sidebar from './Sidebar.svelte';
	import SkipLink from './SkipLink.svelte';
	import { navigationFor, type OrganizationType } from './navigation';

	interface Props {
		organizationType: OrganizationType;
		currentPath: string;
		children: Snippet;
	}

	let { organizationType, currentPath, children }: Props = $props();

	const navigation = $derived(navigationFor(organizationType));
</script>

<SkipLink />
<div class="flex min-h-dvh">
	<Sidebar items={[...navigation.primary, ...navigation.more]} {currentPath} />
	<main
		id="konten"
		tabindex="-1"
		class="mx-auto w-full max-w-content flex-1 px-4 pt-6 pb-(--toast-offset) lg:px-8 lg:pb-8"
	>
		{@render children()}
	</main>
</div>
<BottomNav primary={navigation.primary} more={navigation.more} {currentPath} />
