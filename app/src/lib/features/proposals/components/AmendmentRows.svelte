<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import { proposals as id } from '#lib/i18n/proposals.js';
	import { changedRows, type AmendmentRow } from '../amendment.svelte';
	import type { InvitationErrors } from '../schema';
	import ItemTermsRow from './ItemTermsRow.svelte';

	interface Props {
		rows: AmendmentRow[];
		errors: InvitationErrors;
		onremove: (productId: string) => void;
	}

	let { rows, errors, onremove }: Props = $props();

	const text = id.proposals;
	const modeOptions = Object.entries(text.modes).map(([value, label]) => ({ value, label }));
	const requestIndexOf = (row: AmendmentRow) => changedRows(rows).indexOf(row);
</script>

{#each rows as row (row.item.productId)}
	<div class="flex flex-col gap-2">
		{#if row.mode === 'ADD'}
			<ItemTermsRow item={row.item} index={requestIndexOf(row)} {errors} />
			<Button
				variant="secondary"
				onclick={() => {
					onremove(row.item.productId);
				}}
			>
				{text.removeProduct(row.item.name)}
			</Button>
		{:else}
			<Select label={text.modeLabel(row.item.name)} options={modeOptions} bind:value={row.mode} />
			{#if row.mode === 'UPDATE'}
				<ItemTermsRow item={row.item} index={requestIndexOf(row)} {errors} />
			{/if}
		{/if}
	</div>
{/each}
