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
	<ul class="flex flex-col gap-4">
		{#each demoAccounts as account (account.email)}
			<li class="flex flex-col gap-2">
				<p class="text-sm text-text-muted">
					{text.email}: <span class="font-semibold text-text">{account.email}</span> · {text.password}:
					<span class="font-semibold text-text">{account.password}</span>
				</p>
				<Button
					variant="secondary"
					onclick={() => {
						onpick(account);
					}}>{text[account.role]}</Button
				>
			</li>
		{/each}
	</ul>
</section>
