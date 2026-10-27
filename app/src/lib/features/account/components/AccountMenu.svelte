<script lang="ts">
	import UserRound from '@lucide/svelte/icons/user-round';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import { common as id } from '#lib/i18n/common.js';
	import SessionEnd from './SessionEnd.svelte';

	const text = id.account;
</script>

<div class="group relative">
	<button
		type="button"
		aria-label={text.openMenu}
		aria-haspopup="true"
		class="inline-flex min-h-11 min-w-11 cursor-pointer items-center justify-center gap-2 rounded-pill border border-border bg-surface px-2 font-semibold text-text md:px-4"
	>
		<UserRound class="size-4 text-accent" aria-hidden="true" /><span class="sr-only md:not-sr-only"
			>{text.menu}</span
		><ChevronDown class="hidden size-4 text-text-muted md:block" aria-hidden="true" />
	</button>
	<div
		class="absolute right-0 z-10 hidden pt-2 group-focus-within:block group-hover:block group-has-[dialog[open]]:block"
	>
		<nav
			aria-label={text.menu}
			class="flex w-max flex-col gap-2 rounded-card border border-border bg-surface p-2 shadow-raised"
		>
			<a href="/app/pengaturan" class="secondary-link">{text.manageProfile}</a>
			<SessionEnd />
		</nav>
	</div>
</div>
