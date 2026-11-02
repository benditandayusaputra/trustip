<script lang="ts">
	import Bell from '@lucide/svelte/icons/bell';
	import { afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { apiClient } from '#lib/api/client.js';
	import { openNotificationStream } from '#lib/api/stream.js';
	import { isCurrent } from '#lib/components/layout/navigation.js';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { safeHref } from '#lib/domain/safe-href.js';
	import { common as id } from '#lib/i18n/common.js';
	import {
		findNotification,
		notificationArrivedEvent,
		readNotification,
		UnreadCounter,
		unreadChangedEvent
	} from '../inbox.svelte';
	import { LiveNotifications, watchVisibility } from '../live.svelte';

	interface Props {
		streamUrl: string;
	}

	let { streamUrl }: Props = $props();

	const counter = new UnreadCounter(apiClient);
	const toaster = getToaster();
	const badgeLimit = 99;
	const inboxHref = '/app/pemberitahuan';
	const current = $derived(isCurrent(inboxHref, page.url.pathname));

	async function announce(notificationId: string) {
		const notification = await findNotification(apiClient(), notificationId).catch(() => undefined);
		window.dispatchEvent(new Event(notificationArrivedEvent));
		const href = notification ? safeHref(notification.link) : undefined;
		const detail = notification?.body === '' ? notification.title : notification?.body;
		toaster.show({
			tone: 'info',
			message: detail ?? id.notifications.arrived,
			action:
				notification && href
					? {
							label: id.notifications.open,
							href,
							onclick: () => {
								void readNotification(apiClient(), notification.id).catch(() => undefined);
							}
						}
					: { label: id.notifications.open, href: inboxHref }
		});
	}

	afterNavigate(({ type }) => {
		if (type !== 'enter') {
			void counter.refresh();
		}
	});

	$effect(() => {
		void counter.refresh();
		const recount = () => {
			void counter.refresh();
		};
		window.addEventListener(unreadChangedEvent, recount);
		const live = new LiveNotifications({
			connect: (handlers) => openNotificationStream(apiClient(), streamUrl, handlers),
			refresh: () => counter.refresh(),
			onarrive: (notification) => {
				void counter.refresh();
				void announce(notification.id);
			},
			isVisible: () => document.visibilityState === 'visible'
		});
		live.start();
		const unwatch = watchVisibility(live);
		return () => {
			window.removeEventListener(unreadChangedEvent, recount);
			unwatch();
			live.stop();
		};
	});
</script>

<a
	href={inboxHref}
	aria-label={id.notifications.bell(counter.count)}
	aria-current={current ? 'page' : undefined}
	class="relative inline-flex size-11 items-center justify-center rounded-control {current
		? 'bg-accent-surface text-accent'
		: 'text-text hover:bg-surface-hover'}"
>
	<Bell class="size-6" aria-hidden="true" />
	{#if counter.count > 0}
		<span
			aria-hidden="true"
			class="absolute top-0 right-0 min-w-5 rounded-pill bg-danger-fill px-1 text-center text-xs leading-5 font-semibold text-on-accent ring-2 ring-surface"
		>
			{counter.count > badgeLimit ? `${String(badgeLimit)}+` : counter.count}
		</span>
	{/if}
</a>
