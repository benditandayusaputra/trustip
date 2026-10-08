<script lang="ts">
	import { onMount } from 'svelte';
	import { apiClient } from '#lib/api/client.js';
	import { watchOnline } from '#lib/offline/network.js';
	import { refreshWorkPackets } from '#lib/offline/work-packet.js';

	onMount(() =>
		watchOnline((online) => {
			if (online) {
				void refreshWorkPackets(apiClient());
			}
		})
	);
</script>
