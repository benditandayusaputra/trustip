<script lang="ts">
	import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
	import { common as id } from '#lib/i18n/common.js';
	import { getAccountSession } from '../session.svelte';

	const session = getAccountSession();
</script>

<div class="flex flex-wrap items-center justify-between gap-4">
	<div class="flex flex-col gap-2">
		<p class="section-label">{id.layout.overview}</p>
		<h1 class="text-xl font-semibold text-text">
			{id.account.greeting(session.account?.full_name ?? '')}
		</h1>
		<p class="text-sm text-text-muted">{id.layout.dashboardIntro}</p>
	</div>
	<a
		href={session.account?.membership?.organization_type === 'STORE'
			? '/app/hitung/baru'
			: '/cari/toko'}
		class="action-link text-sm"
	>
		{session.account?.membership?.organization_type === 'STORE'
			? id.nav.count
			: id.layout.searchPartners}
		<ArrowUpRight class="size-4" aria-hidden="true" />
	</a>
</div>
