<script lang="ts">
	import UserRound from '@lucide/svelte/icons/user-round';
	import type { Account } from '#lib/api/users.js';
	import { common as id } from '#lib/i18n/common.js';

	let { account }: { account: Account } = $props();

	const text = id.account;
</script>

<section
	aria-labelledby="akun-nama"
	class="surface-card grid grid-cols-1 items-center gap-6 p-5 compact:p-6 md:grid-cols-(--profile-columns)"
>
	<div class="flex flex-col items-center gap-2 text-center">
		<span
			class="profile-avatar flex items-center justify-center rounded-pill bg-accent-surface text-accent"
			aria-hidden="true"><UserRound class="size-12" /></span
		>
		<h2 id="akun-nama" class="text-lg font-semibold text-text">{account.full_name}</h2>
		{#if account.membership}
			<p class="text-sm text-text-muted">{text.roles[account.membership.role]}</p>
		{/if}
	</div>
	<dl class="grid grid-cols-1 gap-4 compact:grid-cols-2">
		<div class="flex min-w-0 flex-col gap-1">
			<dt class="text-sm font-medium text-text-muted">{text.fullName}</dt>
			<dd class="profile-value">{account.full_name}</dd>
		</div>
		<div class="flex min-w-0 flex-col gap-1">
			<dt class="text-sm font-medium text-text-muted">{text.email}</dt>
			<dd class="profile-value">{account.email}</dd>
		</div>
	</dl>
</section>
