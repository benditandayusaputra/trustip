<script lang="ts">
	import TextField from '#lib/components/ui/TextField.svelte';
	import { directory as id } from '#lib/i18n/directory.js';
	import type { StoreQuery } from '../query';
	interface Props {
		query: StoreQuery;
		onchange: (changes: Partial<StoreQuery>) => void;
	}
	let { query, onchange }: Props = $props();
	let typed = $derived(query.text);
	const text = id.directory;
	const typingDelayMs = 300;
	const maxSearchLength = 80;
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

<TextField label={text.search} type="search" maxlength={maxSearchLength} bind:value={typed} />
