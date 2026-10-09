<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
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

	interface Props {
		placement: Placement;
		onsignedout: () => void;
		onmisplaced: (account: Account) => void;
		actions?: Snippet;
		children: Snippet;
	}

	let { placement, onsignedout, onmisplaced, actions, children }: Props = $props();

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
	const isVisitor = $derived(session.status === 'signedOut' && placement === 'anyone');
	const isOpenWhileLoading = $derived(session.status === 'loading' && placement === 'anyone');
	const signInHref = $derived(
		`/masuk?${new URLSearchParams({ lanjut: page.url.pathname + page.url.search }).toString()}`
	);

	$effect(() => {
		if (session.status === 'signedOut' && placement !== 'anyone') {
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
			<a href={isVisitor ? '/' : '/app'} class="text-lg font-semibold text-text">{id.brand}</a>
			{#if isVisitor}
				<a
					href={signInHref}
					class="inline-flex min-h-11 items-center rounded-control border border-border-strong px-4 font-semibold"
				>
					{id.nav.signIn}
				</a>
			{:else if session.status === 'ready'}
				<div class="flex flex-wrap items-center gap-2">
					{#if isPlaced && actions}
						{@render actions()}
					{/if}
					{#await import('./AccountMenu.svelte') then { default: AccountMenu }}
						<AccountMenu />
					{/await}
				</div>
			{/if}
		</div>
	</header>
	<main id="konten" tabindex="-1" class="mx-auto w-full max-w-content flex-1 px-4 py-6">
		{#if isPlaced || isVisitor || isOpenWhileLoading}
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
