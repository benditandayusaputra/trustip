<script lang="ts">
	import { onMount } from 'svelte';
	import { apiClient } from '#lib/api/client.js';
	import { outboxScheduler } from '#lib/offline/device.js';
	import { watchOnline } from '#lib/offline/network.js';
	import { refreshWorkPackets } from '#lib/offline/work-packet.js';

	const scheduler = outboxScheduler();

	function catchUp() {
		void scheduler.run();
		void refreshWorkPackets(apiClient());
	}

	function whenIdle(work: () => void) {
		if ('requestIdleCallback' in window) {
			window.requestIdleCallback(work);
			return;
		}
		setTimeout(work);
	}

	function sendWhenVisible() {
		if (document.visibilityState === 'visible') {
			void scheduler.run();
		}
	}

	onMount(() => {
		const stopWatching = watchOnline((online) => {
			if (online) {
				whenIdle(catchUp);
			}
		});
		document.addEventListener('visibilitychange', sendWhenVisible);
		return () => {
			stopWatching();
			document.removeEventListener('visibilitychange', sendWhenVisible);
			scheduler.stop();
		};
	});
</script>
