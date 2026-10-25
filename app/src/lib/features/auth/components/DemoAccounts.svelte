<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import { common as id } from '#lib/i18n/common.js';
	import { demoAccounts, type DemoAccount } from '../demo.js';

	interface Props {
		onpick: (account: DemoAccount) => void;
	}

	let { onpick }: Props = $props();

	const text = id.auth.login.demo;
</script>

<section
	id="coba-demo"
	aria-labelledby="coba-demo-judul"
	class="flex flex-col gap-4 rounded-card border border-border bg-surface p-4"
>
	<h2 id="coba-demo-judul" class="text-lg font-semibold text-text">{text.title}</h2>
	<p class="text-base text-text-muted">{text.body}</p>
	<p class="text-sm text-text-muted">
		{text.password}: <span class="font-semibold text-text">{demoAccounts[0]?.password}</span>
	</p>
	<ul class="flex flex-col gap-2">
		{#each demoAccounts as account (account.email)}
			{@const persona = text.accounts[account.persona]}
			<li class="flex flex-col gap-4 rounded-control border border-border p-4">
				<div class="flex items-center gap-4">
					<span
						class="flex size-11 shrink-0 items-center justify-center rounded-pill bg-accent-surface font-semibold text-accent"
						aria-hidden="true">{persona.initial}</span
					>
					<span class="flex min-w-0 flex-col gap-1">
						<span class="font-semibold text-text">{persona.name}</span>
						<span class="text-sm text-text-muted">{persona.role}</span>
						<span class="text-sm wrap-anywhere text-text-muted">
							{text.email}: <span class="font-medium text-text">{account.email}</span>
						</span>
					</span>
				</div>
				<Button
					variant="secondary"
					onclick={() => {
						onpick(account);
					}}>{persona.fill}</Button
				>
			</li>
		{/each}
	</ul>
</section>
