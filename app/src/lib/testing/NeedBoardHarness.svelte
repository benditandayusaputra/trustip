<script lang="ts">
	import NeedBoardView from '#lib/features/directory/components/NeedBoardView.svelte';
	import {
		boardQueryFrom,
		boardSearchParams,
		type BoardQuery
	} from '#lib/features/directory/board-query.js';

	interface Props {
		canRespond?: boolean;
		onaddress?: (address: string) => void;
	}

	let { canRespond = true, onaddress = () => undefined }: Props = $props();

	let query = $state<BoardQuery>(boardQueryFrom(new URLSearchParams()));

	function show(next: BoardQuery) {
		query = next;
		onaddress(boardSearchParams(next).toString());
	}
</script>

<NeedBoardView {query} {canRespond} onchange={show}>
	{#snippet place()}
		<p>{query.origin}</p>
	{/snippet}
</NeedBoardView>
