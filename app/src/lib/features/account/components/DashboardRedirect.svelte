<script lang="ts">
	import { goto } from '$app/navigation';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { dashboardPath, getAccountSession } from '../session.svelte';

	const session = getAccountSession();

	$effect(() => {
		const type = session.account?.membership?.organization_type;
		if (type) {
			void goto(dashboardPath(type), { replaceState: true });
		}
	});
</script>

<Skeleton />
