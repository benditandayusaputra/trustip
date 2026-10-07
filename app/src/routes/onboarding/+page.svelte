<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import { safeNextPath, signInHref } from '#lib/features/auth/links.js';
	import OnboardingWizard from '#lib/features/onboarding/components/OnboardingWizard.svelte';
	import { id } from '#lib/i18n/id.js';

	const next = $derived(safeNextPath(page.url.searchParams.get('lanjut')));
</script>

<svelte:head>
	<title>{id.onboarding.title}</title>
</svelte:head>

<AccountGate
	placement="newcomer"
	onsignedout={() => {
		window.location.assign(signInHref(page.url.pathname + page.url.search));
	}}
	onmisplaced={() => void goto(next, { replaceState: true })}
>
	<section class="mx-auto flex w-full max-w-form flex-col gap-6">
		<h1 class="text-xl font-semibold text-text">{id.onboarding.title}</h1>
		<OnboardingWizard onfounded={() => void goto(next, { replaceState: true })} />
	</section>
</AccountGate>
