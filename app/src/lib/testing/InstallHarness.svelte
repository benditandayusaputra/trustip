<script lang="ts">
	import InstallPrompt from '#lib/features/system/InstallPrompt.svelte';
	import type { InstallEnvironment } from '#lib/features/system/install.svelte.js';

	interface Props {
		environment: InstallEnvironment;
	}

	let { environment }: Props = $props();
</script>

<InstallPrompt {environment} />
