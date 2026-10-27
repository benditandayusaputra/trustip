<script lang="ts">
	import { common as id } from '#lib/i18n/common.js';
	import { demoAccounts, type DemoAccount } from '../demo.js';

	interface Props {
		onpick: (account: DemoAccount) => void;
	}

	let { onpick }: Props = $props();

	const text = id.auth.login.demo;
</script>

<section id="coba-demo" aria-labelledby="coba-demo-judul" class="flex flex-col gap-2">
	<h2 id="coba-demo-judul" class="text-lg font-semibold text-text">{text.title}</h2>
	<p class="text-sm text-text">
		{text.password}: <span class="font-semibold text-text">{demoAccounts[0]?.password}</span>
	</p>
	<ul class="flex flex-col gap-4 pt-2">
		{#each demoAccounts as account (account.email)}
			{@const persona = text.accounts[account.persona]}
			<li
				class="flex flex-col gap-4 rounded-control border border-border p-4 md:flex-row md:items-center md:justify-between"
			>
				<div class="flex min-w-0 items-center gap-4">
					<span
						class="flex size-11 shrink-0 items-center justify-center rounded-pill bg-accent-surface font-semibold text-accent"
						aria-hidden="true">{persona.initial}</span
					>
					<span class="flex min-w-0 flex-col gap-1">
						<span class="font-semibold text-text">{persona.name}</span>
						<span class="text-xs text-text-muted">{persona.role}</span>
						<span class="text-xs wrap-anywhere text-text-muted">
							{text.email}: <span class="font-medium text-text">{account.email}</span>
						</span>
					</span>
				</div>
				<button
					type="button"
					onclick={() => {
						onpick(account);
					}}
					class="inline-flex min-h-11 items-center justify-center rounded-control border border-border bg-surface px-4 text-sm font-semibold text-text hover:border-accent hover:text-accent md:shrink-0"
					>{persona.fill}</button
				>
			</li>
		{/each}
	</ul>
</section>
