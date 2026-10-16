<script lang="ts">
	import { onMount, tick } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import { watchOnline } from '#lib/offline/network.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { steps } from '../schema';
	import { OnboardingWizard, type SubmitOutcome } from '../wizard.svelte';
	import LocationFields from './LocationFields.svelte';
	import ProfileStep from './ProfileStep.svelte';
	import TypeStep from './TypeStep.svelte';

	const id = { ...commonText, ...organizationText };

	interface Props {
		onfounded: () => void;
	}

	let { onfounded }: Props = $props();

	const text = id.onboarding;
	const wizard = new OnboardingWizard(apiClient);
	const toaster = getToaster();
	let formElement: HTMLFormElement;
	let heading: HTMLHeadingElement;
	let isOnline = $state(true);
	const isLastStep = $derived(wizard.step === 'location');

	onMount(() => {
		void wizard.regions.loadProvinces();
		return watchOnline((online) => (isOnline = online));
	});

	async function moveTo(change: () => void) {
		change();
		await tick();
		heading.focus();
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!isLastStep) {
			const before = wizard.step;
			wizard.next();
			await (wizard.step === before ? focusFirstProblem() : moveTo(() => undefined));
			return;
		}
		report(await wizard.submit());
	}

	async function focusFirstProblem() {
		await tick();
		const invalid = formElement.querySelector<HTMLElement>('[aria-invalid="true"]');
		const target = invalid?.matches('input, select') ? invalid : invalid?.querySelector('input');
		target?.focus();
	}

	function report(outcome: SubmitOutcome) {
		if (outcome.kind === 'founded') {
			toaster.show({ tone: 'success', message: text.founded });
			onfounded();
			return;
		}
		if (outcome.kind === 'invalid') {
			void focusFirstProblem();
			return;
		}
		const retry = {
			label: id.auth.retry,
			onclick: () => {
				formElement.requestSubmit();
			}
		};
		toaster.show({
			tone: 'danger',
			message: outcome.message,
			...(outcome.canRetry && { action: retry })
		});
	}
</script>

<form
	bind:this={formElement}
	onsubmit={handleSubmit}
	novalidate
	class="surface-card flex flex-col gap-5 p-5 compact:p-6"
>
	<div class="flex flex-col gap-1">
		<p class="text-sm text-text-muted">{text.progress(wizard.stepNumber, steps.length)}</p>
		<h2 bind:this={heading} tabindex="-1" class="text-lg font-semibold text-text">
			{text.steps[wizard.step]}
		</h2>
	</div>
	{#if wizard.step === 'type'}
		<TypeStep {wizard} />
	{:else if wizard.step === 'profile'}
		<ProfileStep {wizard} />
	{:else}
		<LocationFields
			regions={wizard.regions}
			values={wizard.values}
			errors={wizard.errors}
			isConsignor={wizard.values.type === 'CONSIGNOR'}
		/>
	{/if}
	{#if isLastStep && !isOnline}
		<p role="status" class="text-sm text-warning">{text.offline}</p>
	{/if}
	<div class="flex flex-wrap justify-between gap-4">
		{#if wizard.stepNumber > 1}
			<Button
				variant="secondary"
				onclick={() =>
					moveTo(() => {
						wizard.back();
					})}>{text.back}</Button
			>
		{/if}
		<Button type="submit" busy={wizard.busy} disabled={isLastStep && !isOnline}>
			{isLastStep ? text.submit : text.next}
		</Button>
	</div>
</form>
