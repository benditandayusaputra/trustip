<script lang="ts">
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import ConsignorDirectoryView from '#lib/features/directory/components/ConsignorDirectoryView.svelte';
	import {
		consignorAddress,
		consignorQueryFrom,
		type ConsignorQuery
	} from '#lib/features/directory/consignor-query.js';

	interface Props {
		address?: string;
		onaddress?: (address: string) => void;
	}

	let { address = '', onaddress = () => undefined }: Props = $props();

	let query = $derived<ConsignorQuery>(consignorQueryFrom(new URLSearchParams(address)));

	function show(next: ConsignorQuery) {
		query = next;
		onaddress(consignorAddress(next).toString());
	}

	setToaster(new ToastQueue());
</script>

<ConsignorDirectoryView {query} onchange={show}>
	{#snippet place()}
		<p>{query.origin}</p>
	{/snippet}
</ConsignorDirectoryView>
