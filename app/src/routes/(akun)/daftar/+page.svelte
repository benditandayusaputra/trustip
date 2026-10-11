<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { PUBLIC_TURNSTILE_SITE_KEY } from '$app/env/public';
	import RegisterForm from '#lib/features/auth/components/RegisterForm.svelte';
	import { signInHref } from '#lib/features/auth/links.js';
	import { common as id } from '#lib/i18n/common.js';

	const text = id.auth.register;
	const signIn = $derived(signInHref(page.url.searchParams.get('lanjut')));
</script>

<svelte:head>
	<title>{text.title}</title>
</svelte:head>

<section class="flex flex-col gap-6 rounded-card border border-border bg-surface p-5 md:p-8">
	<header class="flex flex-col gap-2">
		<h1 class="text-kpi font-semibold text-text">{text.title}</h1>
		<p class="text-base text-text-muted">{text.intro}</p>
	</header>
	<RegisterForm siteKey={PUBLIC_TURNSTILE_SITE_KEY} onregistered={() => void goto(signIn)} />
	<p class="border-t border-border pt-6 text-center text-base text-text-muted">
		{text.hasAccount}
		<a href={signIn} class="inline-flex min-h-11 items-center font-semibold text-accent"
			>{text.signIn}</a
		>
	</p>
</section>
