<script lang="ts" generics="Row extends Record<string, string>">
	interface Column {
		key: Extract<keyof Row, string>;
		label: string;
		numeric?: boolean;
	}

	interface Props {
		caption: string;
		columns: Column[];
		rows: Row[];
		rowKey: Extract<keyof Row, string>;
	}

	let { caption, columns, rows, rowKey }: Props = $props();

	const alignment = (column: Column) =>
		column.numeric ? 'text-right font-mono font-medium tabular-nums' : 'text-left';
</script>

<div
	class="table-surface hidden overflow-x-auto rounded-card border border-border bg-surface md:block"
>
	<table class="w-full border-collapse text-sm text-text">
		<caption class="sr-only">{caption}</caption>
		<thead>
			<tr class="border-b border-border bg-surface-hover text-text-muted">
				{#each columns as column (column.key)}
					<th
						scope="col"
						class="px-4 py-4 font-medium {column.numeric ? 'text-right' : 'text-left'}"
						>{column.label}</th
					>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each rows as row (row[rowKey])}
				<tr class="border-b border-border hover:bg-surface-hover">
					{#each columns as column (column.key)}
						<td class="px-4 py-4 {alignment(column)}">{row[column.key]}</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<ul class="flex flex-col gap-2 md:hidden" aria-label={caption}>
	{#each rows as row (row[rowKey])}
		<li class="rounded-card border border-border bg-surface p-4">
			<dl class="grid grid-cols-2 gap-2">
				{#each columns as column (column.key)}
					<dt class="text-sm text-text-muted">{column.label}</dt>
					<dd class={alignment(column)}>{row[column.key]}</dd>
				{/each}
			</dl>
		</li>
	{/each}
</ul>
