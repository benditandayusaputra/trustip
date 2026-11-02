<script lang="ts">
	import Bell from '@lucide/svelte/icons/bell';
	import { afterNavigate } from '$app/navigation';
	import { apiClient } from '#lib/api/client.js';
	import { openNotificationStream } from '#lib/api/stream.js';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { common as id } from '#lib/i18n/common.js';
	import { UnreadCounter, unreadChangedEvent } from '../inbox.svelte';
	import { LiveNotifications, watchVisibility } from '../live.svelte';

	interface Props {
		streamUrl: string;
	}

	let { streamUrl }: Props = $props();

	const counter = new UnreadCounter(apiClient);
	const toaster = getToaster();
	const badgeLimit = 99;

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
			onarrive: () => {
				void counter.refresh();
				toaster.show({
					tone: 'info',
					message: id.notifications.arrived,
					action: { label: id.notifications.open, href: '/app/pemberitahuan' }
				});
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
	href="/app/pemberitahuan"
	aria-label={id.notifications.bell(counter.count)}
	class="relative inline-flex size-11 items-center justify-center rounded-control text-text"
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
