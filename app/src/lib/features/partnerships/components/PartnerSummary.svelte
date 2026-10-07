<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { Partnership } from '#lib/api/partnerships.js';
	import { whatsAppChatUrl } from '#lib/domain/whatsapp.js';
	import { formatDate } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		partnership: Partnership;
	}

	let { partnership }: Props = $props();

	const text = id.partners;
	const counterpart = $derived(partnership.counterpart);
	const chatUrl = $derived(whatsAppChatUrl(counterpart.whatsapp));
</script>

<div class="flex flex-col gap-4">
	<div class="flex flex-wrap items-center gap-2">
		<Badge
			tone={partnership.status === 'ACTIVE' ? 'success' : 'neutral'}
			label={text.statuses[partnership.status]}
		/>
		<span class="text-sm text-text-muted">
			{text.types[counterpart.type]} · {text.since(formatDate(partnership.started_at))}
			{#if partnership.ended_at}
				· {text.endedOn(formatDate(partnership.ended_at))}
			{/if}
		</span>
	</div>
	{#if counterpart.description}
		<p class="text-base text-text">{counterpart.description}</p>
	{/if}
	<dl class="grid grid-cols-2 gap-2 text-sm">
		<dt class="text-text-muted">{text.area}</dt>
		<dd>{[counterpart.district, counterpart.city_name].filter(Boolean).join(', ')}</dd>
		{#if counterpart.address}
			<dt class="text-text-muted">{text.address}</dt>
			<dd>{counterpart.address}</dd>
		{/if}
	</dl>
	{#if chatUrl}
		<a
			href={chatUrl}
			target="_blank"
			rel="noopener noreferrer"
			class="inline-flex min-h-12 items-center self-start rounded-control border border-border-strong px-5 font-semibold text-text"
		>
			{text.contact}
			<span class="sr-only">{id.auth.opensInNewTab}</span>
		</a>
	{/if}
</div>
