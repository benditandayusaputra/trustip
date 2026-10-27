<script lang="ts">
	import { untrack } from 'svelte';
	import UserRound from '@lucide/svelte/icons/user-round';
	import Button from '#lib/components/ui/Button.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { AccountProfile } from '#lib/api/account.js';
	import { localWhatsApp } from '#lib/domain/whatsapp.js';
	import { common as id } from '#lib/i18n/common.js';
	import { AccountEditor, type AccountOutcome } from '../profile.svelte';
	import PasswordDialog from './PasswordDialog.svelte';

	interface Props {
		account: AccountProfile;
		onsaved: (account: AccountProfile) => void;
	}

	let { account, onsaved }: Props = $props();

	const text = id.account;
	const toaster = getToaster();
	const editor = new AccountEditor(
		apiClient,
		untrack(() => $state.snapshot(account)),
		(saved) => {
			onsaved(saved);
		}
	);
	const avatarSize = 96;
	let editing = $state(false);
	let changingPassword = $state(false);
	let picker: HTMLInputElement;

	function report(outcome: AccountOutcome, message: string): boolean {
		if (outcome.kind === 'saved') {
			toaster.show({ tone: 'success', message });
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		}
		return outcome.kind === 'saved';
	}

	async function pickAvatar() {
		const file = picker.files?.[0];
		picker.value = '';
		if (file) {
			report(await editor.chooseAvatar(file), text.avatarSaved);
		}
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		editing = !report(await editor.save(), text.saved);
	}
</script>

<section
	aria-labelledby="akun-nama"
	class="surface-card grid grid-cols-1 items-start gap-6 p-5 compact:p-6 md:grid-cols-(--profile-columns)"
>
	<div class="flex flex-col items-center gap-2 text-center">
		{#if account.avatar}
			<img
				src={account.avatar.url}
				alt={text.avatarAlt(account.full_name)}
				width={avatarSize}
				height={avatarSize}
				class="profile-avatar rounded-pill border border-border object-cover"
			/>
		{:else}
			<span
				class="profile-avatar flex items-center justify-center rounded-pill bg-accent-surface text-accent"
				aria-hidden="true"><UserRound class="size-12" /></span
			>
		{/if}
		<h2 id="akun-nama" class="text-lg font-semibold text-text">{account.full_name}</h2>
		{#if account.membership}
			<p class="text-sm text-text-muted">{text.roles[account.membership.role]}</p>
		{/if}
		<label
			class="inline-flex min-h-11 cursor-pointer items-center text-sm font-semibold text-accent focus-within:outline-2 focus-within:outline-accent hover:underline"
		>
			{text.changeAvatar}
			<input
				bind:this={picker}
				type="file"
				accept="image/jpeg,image/png,image/webp"
				disabled={editor.busy}
				aria-invalid={editor.errors.avatar ? true : undefined}
				onchange={() => void pickAvatar()}
				class="sr-only"
			/>
		</label>
		{#if editor.errors.avatar}<p role="alert" class="text-xs text-danger">
				{editor.errors.avatar}
			</p>{/if}
		{#if account.avatar}
			<button
				type="button"
				disabled={editor.busy}
				onclick={async () => report(await editor.removeAvatar(), text.avatarRemoved)}
				class="inline-flex min-h-11 items-center text-sm text-text-muted hover:text-danger"
				>{text.removeAvatar}</button
			>
		{/if}
	</div>
	{#if editing}
		<form onsubmit={submit} novalidate class="flex flex-col gap-4">
			<TextField
				label={text.fullName}
				name="full_name"
				bind:value={editor.fullName}
				error={editor.errors.fullName}
			/>
			<TextField
				label={text.whatsapp}
				name="whatsapp"
				type="tel"
				hint={text.whatsappHint}
				bind:value={editor.whatsapp}
				error={editor.errors.whatsapp}
			/>
			<div class="form-footer">
				<Button
					variant="secondary"
					onclick={() => {
						editor.reset();
						editing = false;
					}}>{text.cancel}</Button
				>
				<Button type="submit" busy={editor.busy}>{text.save}</Button>
			</div>
		</form>
	{:else}
		<div class="flex flex-col gap-4">
			<dl class="grid grid-cols-1 gap-4 compact:grid-cols-2">
				<div class="flex min-w-0 flex-col gap-1">
					<dt class="text-sm font-medium text-text-muted">{text.fullName}</dt>
					<dd class="profile-value">{account.full_name}</dd>
				</div>
				<div class="flex min-w-0 flex-col gap-1">
					<dt class="text-sm font-medium text-text-muted">{text.email}</dt>
					<dd class="profile-value">{account.email}</dd>
				</div>
				<div class="flex min-w-0 flex-col gap-1">
					<dt class="text-sm font-medium text-text-muted">{text.whatsapp}</dt>
					<dd class="profile-value">{localWhatsApp(account.whatsapp)}</dd>
				</div>
				<div class="flex min-w-0 flex-col gap-1">
					<dt class="text-sm font-medium text-text-muted">{text.password}</dt>
					<dd class="profile-value justify-between gap-2">
						<span class="text-sm text-text-muted">{text.hiddenPassword}</span>
						<button
							type="button"
							onclick={() => (changingPassword = true)}
							class="inline-flex min-h-11 shrink-0 items-center text-sm font-semibold text-accent hover:underline"
							>{text.changePassword}</button
						>
					</dd>
				</div>
			</dl>
			<div><Button variant="secondary" onclick={() => (editing = true)}>{text.edit}</Button></div>
		</div>
	{/if}
</section>
<PasswordDialog bind:open={changingPassword} {editor} />
