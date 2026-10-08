<script lang="ts">
	import type { ProposalRevisionRecord } from '#lib/api/proposals.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { revisionSummary } from '../history';

	interface Props {
		records: ProposalRevisionRecord[];
	}

	let { records }: Props = $props();

	const text = id.proposals;
</script>

{#if records.length > 1}
	<section aria-labelledby="revision-history" class="flex flex-col gap-4">
		<h2 id="revision-history" class="text-lg font-semibold text-text">{text.history}</h2>
		<ol class="flex flex-col gap-4">
			{#each records.toReversed() as record (record.revision_no)}
				<li class="flex flex-col gap-1 border-l-4 border-l-border-strong pl-4">
					<p class="text-sm text-text-muted">
						{text.revisionBy(
							record.revision_no,
							record.author_name,
							formatDateTime(record.created_at)
						)}
					</p>
					{#each revisionSummary(record) as line (line)}
						<p class="text-base text-text">{line}</p>
					{/each}
					{#if record.message}
						<p class="text-base text-text-muted">{record.message}</p>
					{/if}
				</li>
			{/each}
		</ol>
	</section>
{/if}
