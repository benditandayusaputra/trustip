<script lang="ts">
	import type { Snippet } from 'svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import SkipLink from '#lib/components/layout/SkipLink.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { id } from '#lib/i18n/id.js';
	import type { Account } from '#lib/api/users.js';
	import {
		AccountSession,
		isPlacedFor,
		setAccountSession,
		type Placement
	} from '../session.svelte';
	import AccountMenu from './AccountMenu.svelte';

	interface Props {
		placement: Placement;
		onsignedout: () => void;
		onmisplaced: (account: Account) => void;
		children: Snippet;
	}

	let { placement, onsignedout, onmisplaced, children }: Props = $props();

	const session = new AccountSession(apiClient);
	setAccountSession(session);

	$effect(() => {
		void session.load();
	});

	const isPlaced = $derived(
		session.status === 'ready' &&
			session.account !== undefined &&
			isPlacedFor(session.account, placement)
	);

	$effect(() => {
		if (session.status === 'signedOut') {
			onsignedout();
		}
	});

	$effect(() => {
		if (session.status === 'ready' && session.account && !isPlaced) {
			onmisplaced(session.account);
		}
	});
</script>

<SkipLink />
<div class="flex min-h-dvh flex-col">
	<header class="border-b border-border bg-surface">
		<div class="mx-auto flex max-w-content flex-wrap items-center justify-between gap-4 px-4 py-2">
			<a href="/app" class="text-lg font-semibold text-text">{id.brand}</a>
			{#if session.status === 'ready'}
				<AccountMenu />
			{/if}
		</div>
	</header>
	<main id="konten" tabindex="-1" class="mx-auto w-full max-w-content flex-1 px-4 py-6">
		{#if isPlaced}
			{@render children()}
		{:else if session.status === 'failed'}
			<Alert title={id.account.loadFailed}>
				{session.failure}
				{#snippet action()}
					<Button variant="secondary" onclick={() => session.load()}>{id.auth.retry}</Button>
				{/snippet}
			</Alert>
		{:else}
			<Skeleton />
		{/if}
	</main>
</div>
