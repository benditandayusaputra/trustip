<script lang="ts">
	import Smartphone from '@lucide/svelte/icons/smartphone';
	import Button from '#lib/components/ui/Button.svelte';
	import { id } from '#lib/i18n/id.js';
	import { InstallInvite, browserEnvironment, type InstallEnvironment } from './install.svelte';

	interface Props {
		environment?: InstallEnvironment;
	}

	let { environment = browserEnvironment }: Props = $props();

	const text = id.install;
	const invite = new InstallInvite({
		isInstalled: () => environment.isInstalled(),
		isIos: () => environment.isIos()
	});

	$effect(() => invite.watch());
</script>

{#if invite.kind !== 'hidden'}
	<aside
		aria-label={text.title}
		class="flex flex-wrap items-center gap-4 border-b border-border bg-accent-surface px-4 py-2"
	>
		<Smartphone class="size-5 shrink-0 text-accent" aria-hidden="true" />
		<p class="flex-1 text-sm text-text">
			{invite.kind === 'ios' ? text.iosHint : text.body}
		</p>
		{#if invite.kind === 'prompt'}
			<Button onclick={() => void invite.install()}>{text.install}</Button>
		{/if}
		<Button
			variant="secondary"
			onclick={() => {
				invite.dismiss();
			}}>{text.later}</Button
		>
	</aside>
{/if}
