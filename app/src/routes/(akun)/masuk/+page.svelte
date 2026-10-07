<script lang="ts">
	import { page } from '$app/state';
	import { PUBLIC_TURNSTILE_SITE_KEY } from '$app/env/public';
	import LoginForm from '#lib/features/auth/components/LoginForm.svelte';
	import { registerHref, safeNextPath } from '#lib/features/auth/links.js';
	import { id } from '#lib/i18n/id.js';

	const text = id.auth.login;
	const next = $derived(page.url.searchParams.get('lanjut'));
</script>

<svelte:head>
	<title>{text.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<header class="flex flex-col gap-2">
		<h1 class="text-xl font-semibold text-text">{text.title}</h1>
		<p class="text-base text-text-muted">{text.intro}</p>
	</header>
	<LoginForm
		siteKey={PUBLIC_TURNSTILE_SITE_KEY}
		onsignedin={() => {
			window.location.assign(safeNextPath(next));
		}}
	/>
	<p class="text-base text-text-muted">
		{text.noAccount}
		<a href={registerHref(next)} class="font-semibold text-accent">{text.register}</a>
	</p>
</section>
