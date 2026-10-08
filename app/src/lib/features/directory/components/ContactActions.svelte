<script lang="ts">
	import MessageCircle from '@lucide/svelte/icons/message-circle';
	import Button from '#lib/components/ui/Button.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { getDirectoryContact } from '#lib/api/directory.js';
	import { whatsAppChatUrl } from '#lib/domain/whatsapp.js';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		organizationId: string;
		name: string;
		productId?: string;
		canPropose: boolean;
	}

	let { organizationId, name, productId = '', canPropose }: Props = $props();

	const text = id.directory;
	let chatUrl = $state<string | undefined>();
	let busy = $state(false);
	let failed = $state(false);
	const proposeHref = $derived(
		`/app/ajuan/baru?${new URLSearchParams({
			sasaran: organizationId,
			...(productId === '' ? {} : { produk: productId })
		}).toString()}`
	);

	async function revealChat() {
		busy = true;
		failed = false;
		try {
			const contact = await getDirectoryContact(apiClient(), organizationId);
			chatUrl = whatsAppChatUrl(contact.whatsapp, text.greeting(contact.name));
			failed = chatUrl === undefined;
		} catch {
			failed = true;
		} finally {
			busy = false;
		}
	}
</script>

<div class="flex flex-wrap items-center gap-4">
	{#if canPropose}
		<a
			href={proposeHref}
			class="inline-flex min-h-11 items-center rounded-control bg-accent px-4 font-semibold text-on-accent"
			>{text.propose}</a
		>
	{/if}
	{#if chatUrl}
		<a
			href={chatUrl}
			target="_blank"
			rel="noopener noreferrer"
			class="inline-flex min-h-11 items-center gap-2 font-semibold text-accent"
		>
			<MessageCircle class="size-5" aria-hidden="true" />
			{text.openWhatsApp(name)}
		</a>
	{:else}
		<Button variant="secondary" {busy} onclick={() => void revealChat()}>{text.askWhatsApp}</Button>
	{/if}
	{#if failed}
		<p role="status" class="text-sm text-danger">{text.contactFailed}</p>
	{/if}
</div>
