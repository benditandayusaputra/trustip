<script lang="ts">
	import ShieldAlert from '@lucide/svelte/icons/shield-alert';
	import ShieldCheck from '@lucide/svelte/icons/shield-check';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';

	interface SealEvent {
		event_type: string;
		actor_org_name: string;
		actor_user_name: string;
		occurred_at: string;
		code: string;
	}

	interface Props {
		isValid: boolean;
		events: SealEvent[];
	}

	let { isValid, events }: Props = $props();

	const text = id.seal;
</script>

<section
	class="flex flex-col gap-4 rounded-card border p-4 {isValid
		? 'border-success bg-success-surface'
		: 'border-danger bg-danger-surface'}"
>
	<p
		class="flex items-center gap-2 text-base font-semibold {isValid
			? 'text-success'
			: 'text-danger'}"
	>
		{#if isValid}
			<ShieldCheck class="size-6" aria-hidden="true" />
		{:else}
			<ShieldAlert class="size-6" aria-hidden="true" />
		{/if}
		{isValid ? text.valid : text.invalid}
	</p>
	<p class="text-sm text-text">{isValid ? text.validBody : text.invalidBody}</p>
	{#if events.length > 0}
		<h3 class="text-sm font-medium text-text">{text.events}</h3>
		<ol class="flex flex-col gap-2">
			{#each events as event (event.code)}
				<li class="flex flex-col gap-1 text-sm">
					<span class="text-text">{text.eventNames[event.event_type] ?? event.event_type}</span>
					<span class="text-text-muted">
						{text.by(event.actor_user_name, event.actor_org_name)} · {formatDateTime(
							event.occurred_at
						)}
					</span>
					<span class="text-text">
						{text.code}:
						<a href="/verifikasi/{event.code}" class="font-mono text-accent">{event.code}</a>
					</span>
				</li>
			{/each}
		</ol>
	{/if}
</section>
