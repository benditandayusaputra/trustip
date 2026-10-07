<script lang="ts">
	import type { Snippet } from 'svelte';
	import Dialog from './Dialog.svelte';

	interface Props {
		open?: boolean;
		title: string;
		children: Snippet;
		actions?: Snippet;
	}

	let { open = $bindable(false), title, children, actions }: Props = $props();
</script>

<Dialog bind:open {title} {children} {actions} placement="bottom" />
