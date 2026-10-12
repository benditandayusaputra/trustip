<script lang="ts">
	import { goto } from '$app/navigation';
	import PageSkeleton from '#lib/components/layout/PageSkeleton.svelte';
	import { dashboardPath, getAccountSession } from '../session.svelte';

	const session = getAccountSession();
	const type = $derived(session.account?.membership?.organization_type);
	const target = $derived(type ? dashboardPath(type) : '/app');

	$effect(() => {
		if (type) {
			void goto(target, { replaceState: true });
		}
	});
</script>

<PageSkeleton path={target} />
