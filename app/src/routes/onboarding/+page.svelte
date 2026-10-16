<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Building2 from '@lucide/svelte/icons/building-2';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import { safeNextPath, signInHref } from '#lib/features/auth/links.js';
	import OnboardingWizard from '#lib/features/organization/components/OnboardingWizard.svelte';
	import { organization as id } from '#lib/i18n/organization.js';

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
	<section class="trade-panel mx-auto flex w-full max-w-form flex-col gap-6">
		<PageHeading icon={Building2}>{id.onboarding.title}</PageHeading>
		<OnboardingWizard onfounded={() => void goto(next, { replaceState: true })} />
	</section>
</AccountGate>
