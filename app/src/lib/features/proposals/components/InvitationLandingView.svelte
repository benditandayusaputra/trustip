<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { proposals as id } from '#lib/i18n/proposals.js';
	import { InvitationLanding } from '../record.svelte';

	interface Props {
		code: string;
		signInHref: string;
		registerHref: string;
		onboardingHref: string;
		onclaimed: (proposal: Proposal) => void;
	}

	let { code, signInHref, registerHref, onboardingHref, onclaimed }: Props = $props();

	const text = id.invitation;
	const landing = new InvitationLanding(apiClient);
	const linkClass = 'inline-flex min-h-12 items-center rounded-control px-5 font-semibold';

	$effect(() => {
		void landing.load(code);
	});

	async function open() {
		const proposal = await landing.claim(code);
		if (proposal) {
			onclaimed(proposal);
		}
	}
</script>

{#if landing.status === 'loading'}
	<Skeleton />
{:else if landing.status === 'signedOut'}
	<p class="text-base text-text">{text.signedOutBody}</p>
	<div class="flex flex-wrap gap-2">
		<a href={signInHref} class="{linkClass} bg-primary text-on-accent">{text.signIn}</a>
		<a href={registerHref} class="{linkClass} border border-border-strong text-text"
			>{text.register}</a
		>
	</div>
{:else if landing.status === 'newcomer'}
	<p class="text-base text-text">{text.newcomerBody}</p>
	<a href={onboardingHref} class="{linkClass} bg-primary text-on-accent">{text.onboard}</a>
{:else if landing.status === 'refused' || !landing.invitation}
	<Alert title={text.loadFailed}>{landing.message}</Alert>
{:else}
	{@const invitation = landing.invitation}
	<p class="text-lg text-text">{text.summary(invitation.sender_name)}</p>
	{#if invitation.item_count > 0}
		<p class="text-base text-text-muted">{text.productCount(invitation.item_count)}</p>
	{/if}
	<p class="text-sm text-text-muted">
		{id.proposals.expiresAt(formatDateTime(invitation.expires_at))}
	</p>
	<Button busy={landing.busy} onclick={() => void open()}>{text.open}</Button>
{/if}
