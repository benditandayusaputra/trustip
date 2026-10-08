<script lang="ts">
	import DateField from '#lib/components/ui/DateField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { DropFilter } from '#lib/api/drops.js';
	import { id } from '#lib/i18n/id.js';
	import { PartnerOptions } from '../partners.svelte';

	interface Props {
		filter: DropFilter;
		showPartner: boolean;
		onchange: (filter: DropFilter) => void;
	}

	let { filter, showPartner, onchange }: Props = $props();

	const text = id.drops;
	const partners = new PartnerOptions(apiClient);
	const statusOptions = [
		{ value: '', label: text.allStatuses },
		...Object.entries(text.statuses).map(([value, label]) => ({ value, label }))
	];
	const partnerOptions = $derived([
		{ value: '', label: text.allPartners },
		...partners.items.map((partner) => ({ value: partner.id, label: partner.counterpart.name }))
	]);

	$effect(() => {
		if (showPartner) {
			void partners.load('');
		}
	});

	function change(changes: Partial<DropFilter>) {
		onchange({ ...filter, ...changes });
	}
</script>

<fieldset class="grid grid-cols-1 gap-4 md:grid-cols-2">
	<legend class="sr-only">{text.filters}</legend>
	{#if showPartner}
		<Select
			label={text.partner}
			options={partnerOptions}
			value={filter.partnershipId}
			onchange={(partnershipId: string) => {
				change({ partnershipId });
			}}
		/>
	{/if}
	<Select
		label={text.status}
		options={statusOptions}
		value={filter.status}
		onchange={(status: string) => {
			change({ status: status as DropFilter['status'] });
		}}
	/>
	<DateField
		label={text.from}
		value={filter.from}
		onchange={(from: string) => {
			change({ from });
		}}
	/>
	<DateField
		label={text.to}
		value={filter.to}
		onchange={(to: string) => {
			change({ to });
		}}
	/>
</fieldset>
