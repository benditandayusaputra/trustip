<script lang="ts">
	import { id } from '#lib/i18n/id.js';
	import type { OnboardingWizard } from '../wizard.svelte';

	interface Props {
		wizard: OnboardingWizard;
	}

	let { wizard }: Props = $props();

	const text = id.onboarding;
	const choices = ['STORE', 'CONSIGNOR'] as const;
	const groupId = $props.id();
</script>

<div
	role="radiogroup"
	aria-label={text.steps.type}
	aria-invalid={wizard.errors.type ? true : undefined}
	aria-describedby="{groupId}-hint{wizard.errors.type ? ` ${groupId}-error` : ''}"
	class="flex flex-col gap-4"
>
	{#each choices as choice (choice)}
		<label
			class="flex min-h-11 cursor-pointer items-start gap-4 rounded-control border border-border-strong bg-bg p-4 has-checked:border-accent"
		>
			<input
				type="radio"
				name="type"
				value={choice}
				bind:group={wizard.values.type}
				class="mt-1 size-6 shrink-0 accent-accent"
			/>
			<span class="flex flex-col gap-1">
				<span class="text-base font-semibold text-text">{text.types[choice].label}</span>
				<span class="text-sm text-text-muted">{text.types[choice].body}</span>
			</span>
		</label>
	{/each}
	<p id="{groupId}-hint" class="text-xs text-text-muted">{text.typeFixed}</p>
	{#if wizard.errors.type}
		<p id="{groupId}-error" class="text-sm text-danger">{wizard.errors.type}</p>
	{/if}
</div>
