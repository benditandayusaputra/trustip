<script lang="ts">
	import TextField from '#lib/components/ui/TextField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import { directory as id } from '#lib/i18n/directory.js';
	import { sortChoices, type StoreQuery } from '../query';
	interface Props {
		query: StoreQuery;
		onchange: (changes: Partial<StoreQuery>) => void;
	}
	let { query, onchange }: Props = $props();
	let typed = $derived(query.text);
	const text = id.directory;
	const typingDelayMs = 300;
	const maxSearchLength = 80;
	const sorts = sortChoices.map((value) => ({ value, label: text.sorts[value] }));
	$effect(() => {
		const next = typed;
		const timer = setTimeout(() => {
			if (next !== query.text) onchange({ text: next });
		}, typingDelayMs);
		return () => {
			clearTimeout(timer);
		};
	});
</script>

<div class="grid gap-4 md:grid-cols-3">
	<div class="md:col-span-2">
		<TextField label={text.search} type="search" maxlength={maxSearchLength} bind:value={typed} />
	</div>
	<Select
		label={text.sort}
		options={sorts}
		value={query.sort}
		onchange={(sort: string) => {
			onchange({ sort: sort as StoreQuery['sort'] });
		}}
	/>
</div>
