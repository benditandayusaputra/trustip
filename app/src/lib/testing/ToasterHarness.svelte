<script lang="ts">
	import { untrack } from 'svelte';
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { setToaster, type Toaster as ToasterState } from '#lib/components/ui/toast.svelte.js';

	interface Props {
		toaster: ToasterState;
	}

	let { toaster }: Props = $props();

	setToaster(untrack(() => toaster));
</script>

<Toaster />
