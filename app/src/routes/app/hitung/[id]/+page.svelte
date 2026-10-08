<script lang="ts">
	import { page } from '$app/state';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import CountView from '#lib/features/counts/components/CountView.svelte';
	import { id } from '#lib/i18n/id.js';

	const session = getAccountSession();
	const isStore = $derived(session.account?.membership?.organization_type === 'STORE');
</script>

<svelte:head>
	<title>{id.counts.recapTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.counts.recapTitle}</h1>
	{#key page.params.id}
		<CountView countId={page.params.id ?? ''} {isStore} />
	{/key}
</section>
