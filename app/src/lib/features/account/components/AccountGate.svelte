<script lang="ts">
	import type { Snippet } from 'svelte';
	import { navigating, page } from '$app/state';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import PageSkeleton from '#lib/components/layout/PageSkeleton.svelte';
	import { pendingPagePath } from '#lib/components/layout/page-skeleton.js';
	import AppShell from '#lib/components/layout/AppShell.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { common as id } from '#lib/i18n/common.js';
	import { dashboardDestination } from '#lib/domain/dashboard-access.js';
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
		loading?: Snippet;
		children: Snippet;
	}

	let { placement, onsignedout, onmisplaced, actions, children, loading }: Props = $props();

	const session = new AccountSession(apiClient);
	setAccountSession(session);

	$effect(() => {
		void session.load();
	});

	const currentRoute = $derived(page.route.id ?? page.url.pathname);
	const isPlaced = $derived(
		session.status === 'ready' &&
			session.account !== undefined &&
			isPlacedFor(session.account, placement) &&
			(placement !== 'member' ||
				dashboardDestination(currentRoute, session.account.membership?.organization_type ?? '') ===
					currentRoute)
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

{#snippet toolbar()}
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
{/snippet}

{#snippet content()}
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
		{#if loading}{@render loading()}{:else}<PageSkeleton path={page.url.pathname} />{/if}
	{/if}
{/snippet}

<AppShell
	organizationType={isPlaced && placement !== 'newcomer'
		? session.account?.membership?.organization_type
		: undefined}
	currentPath={page.url.pathname}
	workspace={placement === 'member'}
	pendingPath={pendingPagePath(navigating.to?.url, page.url)}
	actions={toolbar}
	children={content}
/>
