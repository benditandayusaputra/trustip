<script lang="ts">
	import NeedList from '#lib/features/needs/components/NeedList.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.needs.listTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<h1 class="text-xl font-semibold text-text">{id.needs.listTitle}</h1>
		<a href="/app/kebutuhan/baru" class="font-semibold text-accent">{id.needs.post}</a>
	</div>
	<NeedList />
</section>
