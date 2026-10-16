<script lang="ts">
	import Plus from '@lucide/svelte/icons/plus';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Handshake from '@lucide/svelte/icons/handshake';
	import PartnerList from '#lib/features/partnerships/components/PartnerList.svelte';
	import { partnerships as id } from '#lib/i18n/partnerships.js';
</script>

<svelte:head>
	<title>{id.partners.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<PageHeading icon={Handshake}>{id.partners.title}</PageHeading>
		<a
			href="/app/ajuan/baru"
			class="inline-flex min-h-12 items-center gap-2 rounded-control bg-primary px-5 font-semibold text-on-accent"
		>
			<Plus class="size-5" aria-hidden="true" />
			{id.partners.invite}
		</a>
	</div>
	<PartnerList />
</section>
