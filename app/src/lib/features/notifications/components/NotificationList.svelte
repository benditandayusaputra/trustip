<script lang="ts">
	import { goto } from '$app/navigation';
	import Bell from '@lucide/svelte/icons/bell';
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
	<div class="flex flex-col gap-4">
		<div>
			<Button variant="secondary" busy={inbox.busy} onclick={() => void readAll()}>
				{text.markAll}
			</Button>
		</div>
		<ul class="flex flex-col gap-2">
			{#each inbox.items as notification (notification.id)}
				{@const href = safeHref(notification.link) ?? '/app'}
				<li>
					<a
						{href}
						onclick={(event) => void open(event, notification, href)}
						class="flex flex-col gap-1 rounded-control border border-border bg-surface p-4 {notification.is_read
							? ''
							: 'border-l-4 border-l-accent'}"
					>
						<span class="text-base font-semibold text-text">{notification.title}</span>
						<span class="text-sm text-text">{notification.body}</span>
						<span class="text-xs text-text-muted">
							{formatDateTime(notification.created_at)}
							{#if !notification.is_read}
								· {text.unread}
							{/if}
						</span>
					</a>
				</li>
			{/each}
		</ul>
		{#if inbox.nextCursor}
			<Button variant="secondary" busy={inbox.busy} onclick={() => void inbox.loadMore()}>
				{text.loadMore}
			</Button>
		{/if}
	</div>
{/if}
