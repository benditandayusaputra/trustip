<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { listAgreementVersions, type AgreementVersion } from '#lib/api/agreements.js';
	import { formatDate } from '#lib/i18n/format.js';
	import { agreements as id } from '#lib/i18n/agreements.js';
	import { describeItemChange, describeTermChange } from '../history';

	interface Props {
		partnershipId: string;
	}

	let { partnershipId }: Props = $props();

	const text = id.agreement;
	let versions = $state<AgreementVersion[]>();
	let failure = $state('');

	$effect(() => {
		listAgreementVersions(apiClient(), partnershipId)
			.then((loaded) => {
				versions = [...loaded].reverse();
			})
			.catch((error: unknown) => {
				failure = error instanceof Error ? error.message : '';
			});
	});
</script>

<section class="flex flex-col gap-4">
	<h2 class="text-lg font-semibold text-text">{text.history}</h2>
	{#if failure}
		<Alert title={text.historyFailed}>{failure}</Alert>
	{:else if !versions}
		<ContentSkeleton variant="list" />
	{:else}
		<ol class="flex flex-col gap-4">
			{#each versions as version (version.version)}
				<li class="flex flex-col gap-2 rounded-control border border-border bg-surface p-4">
					<span class="flex flex-wrap items-center gap-2">
						<span class="font-semibold text-text">{text.version(version.version)}</span>
						<Badge
							tone={version.status === 'ACTIVE' ? 'success' : 'neutral'}
							label={version.status === 'ACTIVE' ? text.active : text.superseded}
						/>
						<span class="text-sm text-text-muted">
							{version.number} · {formatDate(version.accepted_at)}
						</span>
					</span>
					<ul class="flex list-disc flex-col gap-1 pl-5 text-sm text-text">
						{#each version.term_changes as change (change.field)}
							<li>{describeTermChange(change)}</li>
						{/each}
						{#each version.item_changes as change (change.product_id)}
							<li>{describeItemChange(change)}</li>
						{/each}
					</ul>
					<a
						href="/app/mitra/{partnershipId}/kesepakatan/{version.version}"
						class="text-sm font-semibold text-accent"
					>
						{text.openVersion(version.version)}
					</a>
				</li>
			{/each}
		</ol>
	{/if}
</section>
