<script lang="ts">
	import { PUBLIC_TURNSTILE_SITE_KEY } from '$app/env/public';
	import DemoAccounts from '#lib/features/auth/components/DemoAccounts.svelte';
	import LoginForm from '#lib/features/auth/components/LoginForm.svelte';
	import type { DemoAccount } from '#lib/features/auth/demo.js';
	import { common as id } from '#lib/i18n/common.js';

	const text = id.auth.login;
	let loginForm: { fill: (values: DemoAccount) => void } | undefined = $state();
</script>

<svelte:head>
	<title>{text.demoTitle}</title>
</svelte:head>

<section class="flex flex-col gap-6 rounded-card border border-border bg-surface p-5 md:p-8">
	<header class="flex flex-col gap-2">
		<h1 class="text-xl font-semibold text-text">{text.demoTitle}</h1>
		<p class="text-base text-text-muted">{text.demoIntro}</p>
	</header>
	<DemoAccounts
		onpick={(account: DemoAccount) => {
			loginForm?.fill(account);
		}}
	/>
	<LoginForm
		bind:this={loginForm}
		siteKey={PUBLIC_TURNSTILE_SITE_KEY}
		onsignedin={() => {
			window.location.assign('/app');
		}}
	/>
</section>
