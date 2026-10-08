<script lang="ts">
	import StoreDirectoryView from '#lib/features/directory/components/StoreDirectoryView.svelte';
	import {
		storeQueryFrom,
		storeSearchParams,
		type StoreQuery
	} from '#lib/features/directory/query.js';

	interface Props {
		address?: string;
		onaddress?: (address: string) => void;
	}

	let { address = '', onaddress = () => undefined }: Props = $props();

	let query = $derived<StoreQuery>(storeQueryFrom(new URLSearchParams(address)));

	function show(next: StoreQuery) {
		query = next;
		onaddress(storeSearchParams(next).toString());
	}
</script>

<StoreDirectoryView {query} onchange={show}>
	{#snippet place()}
		<p>{query.origin}</p>
	{/snippet}
</StoreDirectoryView>
