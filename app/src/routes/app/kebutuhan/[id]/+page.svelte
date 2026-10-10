<script lang="ts">
	import { page } from '$app/state';
	import NeedDetail from '#lib/features/needs/components/NeedDetail.svelte';
	import { needs as id } from '#lib/i18n/needs.js';
</script>

<svelte:head>
	<title>{id.needs.listTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<NeedDetail needId={page.params.id ?? ''} />
</section>
