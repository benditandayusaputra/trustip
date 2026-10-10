<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Need } from '#lib/api/needs.js';
	import NeedForm from '#lib/features/needs/components/NeedForm.svelte';
	import { needs as id } from '#lib/i18n/needs.js';
</script>

<svelte:head>
	<title>{id.needs.newTitle}</title>
</svelte:head>

<section class="trade-panel mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.needs.newTitle}</h1>
	<p class="text-base text-text-muted">{id.needs.newIntro}</p>
	<NeedForm onposted={(need: Need) => void goto(`/app/kebutuhan/${need.id}`)} />
</section>
