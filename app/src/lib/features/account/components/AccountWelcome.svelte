<script lang="ts">
	import { id } from '#lib/i18n/id.js';
	import { getAccountSession } from '../session.svelte';

	const session = getAccountSession();
</script>

<svelte:head>
	<title>{id.brand}</title>
</svelte:head>

<section class="flex flex-col gap-2">
	<h1 class="text-xl font-semibold text-text">
		{id.account.greeting(session.account?.full_name ?? '')}
	</h1>
	<p class="text-base text-text-muted">{id.account.nextStep}</p>
</section>
