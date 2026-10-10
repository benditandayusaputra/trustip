<script lang="ts">
	import Copy from '@lucide/svelte/icons/copy';
	import Button from '#lib/components/ui/Button.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';
	import { invitationUrl, senderName, whatsAppShareUrl } from '../links';

	const id = { ...commonText, ...proposalsText };

	interface Props {
		proposal: Proposal;
		inviteCode: string;
	}

	let { proposal, inviteCode }: Props = $props();

	const text = id.proposals;
	const toaster = getToaster();
	const link = $derived(invitationUrl(window.location.origin, inviteCode));
	const shareUrl = $derived(whatsAppShareUrl(text.whatsAppText(senderName(proposal), link)));

	async function copy() {
		await navigator.clipboard.writeText(link);
		toaster.show({ tone: 'success', message: text.copied });
	}
</script>

<section class="flex flex-col gap-4 rounded-card border border-accent bg-surface p-4">
	<h2 class="text-lg font-semibold text-text">{text.shareTitle}</h2>
	<p class="text-sm text-text-muted">{text.shareBody}</p>
	<p class="font-mono text-sm break-all text-text">{link}</p>
	<div class="flex flex-wrap gap-2">
		<Button variant="secondary" onclick={() => void copy()}>
			<Copy class="size-5" aria-hidden="true" />
			{text.copyLink}
		</Button>
		<a
			href={shareUrl}
			target="_blank"
			rel="noopener noreferrer"
			class="inline-flex min-h-12 items-center rounded-control bg-primary px-5 font-semibold text-on-accent"
		>
			{text.shareWhatsApp}
			<span class="sr-only">{id.auth.opensInNewTab}</span>
		</a>
	</div>
</section>
