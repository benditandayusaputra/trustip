<script lang="ts">
	import { untrack } from 'svelte';
	import CircleDashed from '@lucide/svelte/icons/circle-dashed';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Switch from '#lib/components/ui/Switch.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Organization } from '#lib/api/organizations.js';
	import { organization as id } from '#lib/i18n/organization.js';
	import { ListingEditor } from '../listing.svelte';

	interface Props {
		organization: Organization;
		canEdit: boolean;
		onsaved: (organization: Organization) => void;
	}

	let { organization, canEdit, onsaved }: Props = $props();

	const text = id.listing;
	const editor = new ListingEditor(
		apiClient,
		untrack(() => organization)
	);
	const toaster = getToaster();
	const meterId = $props.id();
	const fullPercent = 100;
	const percent = $derived(organization.listing?.percent ?? 0);
	const missing = $derived(organization.listing?.items.filter((item) => !item.is_done) ?? []);

	async function change(isListed: boolean) {
		const outcome = await editor.change(organization.id, isListed);
		if (outcome.kind === 'saved') {
			toaster.show({ tone: 'success', message: isListed ? text.listed : text.hidden });
			onsaved(outcome.organization);
		}
	}
</script>

<section class="surface-card flex flex-col gap-4 p-5 compact:p-6" aria-labelledby="{meterId}-title">
	<h2 id="{meterId}-title" class="text-lg font-semibold text-text">{text.title}</h2>
	<div class="flex flex-col gap-2">
		<label for="{meterId}-meter" class="text-sm font-medium text-text">{text.meter(percent)}</label>
		<progress
			id="{meterId}-meter"
			value={percent}
			max={fullPercent}
			class="h-2 w-full {percent === fullPercent ? 'accent-success' : 'accent-accent'}"
		></progress>
	</div>
	{#if missing.length > 0}
		<ul class="flex flex-col gap-2">
			{#each missing as item (item.requirement)}
				<li class="flex items-start gap-2 text-sm text-text">
					<CircleDashed class="mt-1 size-4 shrink-0 text-text-muted" aria-hidden="true" />
					<span>
						{text.requirements[item.requirement]}
						{#if item.is_required}
							<span class="text-warning">({text.required})</span>
						{/if}
					</span>
				</li>
			{/each}
		</ul>
	{/if}
	{#if canEdit}
		<Switch
			label={text.switchLabel}
			description={text.description}
			busy={editor.busy}
			bind:checked={editor.isListed}
			onchange={(checked: boolean) => void change(checked)}
		/>
	{/if}
	{#if editor.failure}
		<Alert title={text.title}>{editor.failure}</Alert>
	{/if}
</section>
