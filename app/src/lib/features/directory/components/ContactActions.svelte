<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import WhatsAppIcon from '#lib/components/ui/WhatsAppIcon.svelte';
	import { page } from '$app/state';
	import { apiClient } from '#lib/api/client.js';
	import { getDirectoryContact } from '#lib/api/directory.js';
	import { ApiError } from '#lib/api/errors.js';
	import { whatsAppChatUrl } from '#lib/domain/whatsapp.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { publicProfile as publicProfileText } from '#lib/i18n/public-profile.js';

	const id = { ...directoryText, ...publicProfileText };

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
	let needsSignIn = $state(false);
	const signInHref = $derived(
		`/masuk?${new URLSearchParams({ lanjut: page.url.pathname }).toString()}`
	);
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
		} catch (error) {
			needsSignIn = error instanceof ApiError && error.isSessionProblem();
			failed = !needsSignIn;
		} finally {
			busy = false;
		}
	}
</script>

<div class="flex flex-wrap items-center gap-4">
	{#if canPropose}
		<a
			href={proposeHref}
			class="inline-flex min-h-11 items-center rounded-control bg-primary px-4 font-semibold text-on-accent"
			>{text.propose}</a
		>
	{/if}
	{#if chatUrl}
		<a
			href={chatUrl}
			target="_blank"
			rel="noopener noreferrer"
			class="inline-flex min-h-12 items-center gap-2 rounded-control border border-whatsapp bg-whatsapp px-5 font-semibold text-on-accent hover:bg-whatsapp-hover"
		>
			<WhatsAppIcon />
			{text.openWhatsApp(name)}
		</a>
	{:else}
		<Button variant="whatsapp" {busy} onclick={() => void revealChat()}
			><WhatsAppIcon />{text.askWhatsApp}</Button
		>
	{/if}
	{#if needsSignIn}
		<a href={signInHref} class="inline-flex min-h-11 items-center font-semibold text-accent">
			{id.publicProfile.signInToContact}
		</a>
	{/if}
	{#if failed}
		<p role="status" class="text-sm text-danger">{text.contactFailed}</p>
	{/if}
</div>
