<script lang="ts">
	import { goto } from '$app/navigation';
	import Bell from '@lucide/svelte/icons/bell';
	import CheckCheck from '@lucide/svelte/icons/check-check';
	import Clock from '@lucide/svelte/icons/clock';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Notification } from '#lib/api/notifications.js';
	import { safeHref } from '#lib/domain/safe-href.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { common as id } from '#lib/i18n/common.js';
	import { Inbox } from '../inbox.svelte';

	const text = id.notifications;
	const inbox = new Inbox(apiClient);
	const toaster = getToaster();
	let onlyUnread = $state(false);
	const shown = $derived(onlyUnread ? inbox.items.filter((item) => !item.is_read) : inbox.items);
	const filters = $derived([
		{ unread: false, label: text.all, count: inbox.items.length },
		{ unread: true, label: text.unread, count: inbox.unread }
	]);

	$effect(() => {
		void inbox.load();
	});

	async function open(event: MouseEvent, notification: Notification, href: string) {
		event.preventDefault();
		await inbox.open(notification).catch(() => undefined);
		await goto(href);
	}

	async function readAll() {
		await inbox.readAll();
		toaster.show({ tone: 'success', message: text.markedAll });
	}
</script>

{#if inbox.status === 'loading'}
	<ContentSkeleton variant="list" />
{:else if inbox.status === 'failed'}
	<Alert title={text.loadFailed}>
		{inbox.failure}
		{#snippet action()}
			<Button variant="secondary" onclick={() => void inbox.load()}>{id.auth.retry}</Button>
		{/snippet}
	</Alert>
{:else if inbox.items.length === 0}
	<EmptyState icon={Bell} message={text.empty}>
		{#snippet action()}
			<a href="/app" class="font-semibold text-accent">{text.backToDashboard}</a>
		{/snippet}
	</EmptyState>
{:else}
	<div class="flex flex-col gap-6">
		<div class="flex flex-wrap items-center justify-between gap-4">
			<div class="flex flex-wrap items-center gap-2">
				<p class="text-xl font-semibold text-text" aria-hidden="true">{text.title}</p>
				{#if inbox.unread > 0}
					<span
						class="rounded-pill border border-accent bg-accent-surface px-2 text-xs font-semibold text-accent"
						>{text.unreadCount(inbox.unread)}</span
					>
				{/if}
			</div>
			{#if inbox.unread > 0}
				<Button variant="secondary" busy={inbox.busy} onclick={() => void readAll()}>
					<CheckCheck class="size-4" aria-hidden="true" />{text.markAll}
				</Button>
			{/if}
		</div>
		<div role="group" aria-label={text.filter} class="flex gap-4 border-b border-border">
			{#each filters as filter (filter.label)}
				<button
					type="button"
					aria-pressed={onlyUnread === filter.unread}
					onclick={() => (onlyUnread = filter.unread)}
					class="-mb-px inline-flex min-h-11 items-center gap-2 border-b-2 px-1 text-base {onlyUnread ===
					filter.unread
						? 'border-accent font-semibold text-accent'
						: 'border-transparent text-text-muted hover:text-text'}"
				>
					{filter.label}
					<span
						class="rounded-pill bg-surface-hover px-2 text-xs font-semibold text-text"
						aria-hidden="true">{filter.count}</span
					>
				</button>
			{/each}
		</div>
		{#if shown.length === 0}
			<p class="text-sm text-text-muted">{text.noUnread}</p>
		{/if}
		<ul class="flex flex-col gap-4">
			{#each shown as notification (notification.id)}
				{@const href = safeHref(notification.link) ?? '/app'}
				<li
					class="flex flex-col gap-2 rounded-card border border-border bg-surface p-4 compact:p-5 md:flex-row md:items-start md:gap-4 {notification.is_read
						? ''
						: 'border-l-4 border-l-accent'}"
				>
					<a
						{href}
						onclick={(event) => void open(event, notification, href)}
						class="flex min-w-0 flex-1 flex-col gap-1"
					>
						<span
							class="flex items-center gap-2 text-base text-text {notification.is_read
								? ''
								: 'font-semibold'}"
						>
							{#if !notification.is_read}
								<span class="size-2 shrink-0 rounded-pill bg-accent" aria-hidden="true"></span>
							{/if}
							{notification.title}
						</span>
						<span class="text-sm {notification.is_read ? 'text-text-muted' : 'text-text'}"
							>{notification.body}</span
						>
						<span
							class="flex items-center gap-1 text-xs {notification.is_read
								? 'text-text-muted'
								: 'font-semibold text-accent'}"
						>
							<Clock class="size-4 shrink-0" aria-hidden="true" />
							{formatDateTime(notification.created_at)}
							{#if !notification.is_read}
								<span class="sr-only">· {text.unread}</span>
							{/if}
						</span>
					</a>
					{#if !notification.is_read}
						<button
							type="button"
							aria-label={text.markOneLabel(notification.title)}
							onclick={() => void inbox.open(notification)}
							class="inline-flex min-h-11 shrink-0 items-center self-start text-sm text-accent hover:underline"
							>{text.markOne}</button
						>
					{/if}
				</li>
			{/each}
		</ul>
		{#if inbox.nextCursor}
			<div class="flex justify-center">
				<Button variant="secondary" busy={inbox.busy} onclick={() => void inbox.loadMore()}>
					{text.loadMore}
				</Button>
			</div>
		{/if}
	</div>
{/if}
