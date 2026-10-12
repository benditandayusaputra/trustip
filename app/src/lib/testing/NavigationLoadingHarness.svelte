<script lang="ts">
	import NavigationContent from '#lib/components/layout/NavigationContent.svelte';
	let waiting = $state(false);
</script>

<button onclick={() => (waiting = !waiting)}>{waiting ? 'Batalkan' : 'Memuat'}</button>
<NavigationContent pendingPath={waiting ? '/app/laporan' : undefined}>
	<input aria-label="Isian sebelumnya" />
</NavigationContent>
