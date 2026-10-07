<script lang="ts">
	import PartnerList from '#lib/features/partnerships/components/PartnerList.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.partners.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<h1 class="text-xl font-semibold text-text">{id.partners.title}</h1>
		<a
			href="/app/ajuan/baru"
			class="inline-flex min-h-12 items-center rounded-control bg-accent px-5 font-semibold text-on-accent"
		>
			{id.partners.invite}
		</a>
	</div>
	<PartnerList />
</section>
