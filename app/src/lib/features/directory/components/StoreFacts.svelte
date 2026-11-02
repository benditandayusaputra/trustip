<script lang="ts">
	interface Props {
		facts: { label: string; value: string }[];
	}

	let { facts }: Props = $props();
</script>

<dl
	class="grid auto-cols-fr grid-flow-col divide-x divide-border rounded-control border border-border"
>
	{#each facts as fact (fact.label)}
		<div class="flex min-w-0 flex-col-reverse justify-end gap-1 p-2 text-center">
			<dt class="text-xs leading-tight text-text-muted">{fact.label}</dt>
			<dd class="truncate font-mono text-base font-semibold text-text">{fact.value}</dd>
		</div>
	{/each}
</dl>
