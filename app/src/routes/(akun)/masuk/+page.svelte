<script lang="ts">
	import type { Account } from '#lib/api/users.js';
	import { page } from '$app/state';
	import { PUBLIC_TURNSTILE_SITE_KEY } from '$app/env/public';
	import LoginForm from '#lib/features/auth/components/LoginForm.svelte';
	import { registerHref, signedInPath } from '#lib/features/auth/links.js';
	import { common as id } from '#lib/i18n/common.js';

	const text = id.auth.login;
	const next = $derived(page.url.searchParams.get('lanjut'));
</script>

<svelte:head>
	<title>{text.title}</title>
</svelte:head>

<section class="auth-form flex flex-col gap-6">
	<header class="flex flex-col gap-2">
		<h1 class="text-kpi font-semibold text-text">{text.title}</h1>
		<p class="text-base text-text-muted">{text.intro}</p>
	</header>
	<LoginForm
		siteKey={PUBLIC_TURNSTILE_SITE_KEY}
		onsignedin={(account: Account) => {
			window.location.assign(signedInPath(next, account));
		}}
	/>
	<p class="text-center text-base text-text-muted">
		{text.noAccount}
		<a href={registerHref(next)} class="inline-flex min-h-11 items-center font-semibold text-accent"
			>{text.register}</a
		>
	</p>
</section>
