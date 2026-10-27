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

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<header class="flex flex-col gap-2 text-center">
		<h1 class="text-xl font-bold text-text compact:text-kpi-compact">{text.demoTitle}</h1>
		<p class="text-sm text-text-muted">{text.demoIntro}</p>
	</header>
	<div class="auth-form auth-form-wide flex flex-col gap-6">
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
	</div>
</section>
