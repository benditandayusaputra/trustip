<script lang="ts">
	import NumberField from '#lib/components/ui/NumberField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { proposals as id } from '#lib/i18n/proposals.js';
	import type { InvitationErrors, InvitationValues } from '../schema';

	interface Props {
		values: InvitationValues;
		errors: InvitationErrors;
	}

	let { values, errors }: Props = $props();

	const text = id.proposals;
	const maxPaymentTermsDays = 30;
	const options = (labels: Record<string, string>) =>
		Object.entries(labels).map(([value, label]) => ({ value, label }));
</script>

<fieldset class="flex flex-col gap-4">
	<legend class="mb-2 text-lg font-semibold text-text">{text.generalTerms}</legend>
	<Select
		label={text.schedule}
		options={options(text.schedules)}
		bind:value={values.schedule}
		error={errors['terms.settlement_schedule']}
	/>
	<NumberField
		label={text.paymentTerms}
		hint={text.paymentTermsHint}
		max={maxPaymentTermsDays}
		bind:value={values.paymentTermsDays}
		{...errors['terms.payment_terms_days'] && { error: errors['terms.payment_terms_days'] }}
	/>
	<Select
		label={text.rhythm}
		options={options(text.rhythms)}
		bind:value={values.rhythm}
		error={errors['terms.count_rhythm']}
	/>
	<TextField
		label={text.notes}
		hint={text.notesHint}
		bind:value={values.notes}
		error={errors['terms.notes']}
	/>
	<div class="flex flex-col gap-2 rounded-control border border-border p-4">
		<p class="text-sm font-medium text-text">{text.fixedTerms}</p>
		<ul class="flex list-disc flex-col gap-1 pl-5 text-sm text-text-muted">
			{#each text.fixedTermsItems as term (term)}
				<li>{term}</li>
			{/each}
		</ul>
	</div>
</fieldset>
