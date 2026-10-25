<script lang="ts">
	import Plus from '@lucide/svelte/icons/plus';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import NeedList from '#lib/features/needs/components/NeedList.svelte';
	import { needs as id } from '#lib/i18n/needs.js';
</script>

<svelte:head>
	<title>{id.needs.listTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<PageHeading icon={Megaphone}>{id.needs.listTitle}</PageHeading>
		<a href="/app/kebutuhan/baru" class="action-link"
			><Plus class="size-5" aria-hidden="true" />{id.needs.post}</a
		>
	</div>
	<NeedList />
</section>
