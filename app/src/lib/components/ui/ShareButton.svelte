<script lang="ts">
	import Share2 from '@lucide/svelte/icons/share-2';
	import { getToaster } from './toast.svelte';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		title: string;
		url: string;
	}

	let { title, url }: Props = $props();

	const toaster = getToaster();

	async function share() {
		if (typeof navigator.share === 'function') {
			await navigator.share({ title, url }).catch(() => undefined);
			return;
		}
		try {
			await navigator.clipboard.writeText(url);
			toaster.show({ tone: 'success', message: id.share.copied });
		} catch {
			toaster.show({ tone: 'danger', message: id.share.failed });
		}
	}
</script>

<button
	type="button"
	class="inline-flex min-h-11 items-center gap-2 self-start rounded-control border border-border-strong px-4 font-semibold text-text"
	onclick={() => void share()}
>
	<Share2 class="size-5" aria-hidden="true" />
	{id.share.label}
</button>
