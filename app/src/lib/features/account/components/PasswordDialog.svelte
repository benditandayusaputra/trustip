<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import PasswordField from '#lib/components/ui/PasswordField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { common as id } from '#lib/i18n/common.js';
	import type { AccountEditor } from '../profile.svelte';

	let { open = $bindable(false), editor }: { open?: boolean; editor: AccountEditor } = $props();

	const text = id.account;
	const toaster = getToaster();
	let current = $state('');
	let next = $state('');

	function close() {
		current = '';
		next = '';
		editor.errors = {};
		open = false;
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await editor.changePassword(current, next);
		if (outcome.kind === 'saved') {
			toaster.show({ tone: 'success', message: text.passwordSaved });
			close();
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		}
	}
</script>

<Dialog bind:open title={text.passwordTitle}>
	<form onsubmit={submit} novalidate class="flex flex-col gap-4">
		<PasswordField
			label={text.currentPassword}
			name="current_password"
			autocomplete="current-password"
			bind:value={current}
			error={editor.errors.currentPassword}
		/>
		<PasswordField
			label={text.newPassword}
			name="new_password"
			autocomplete="new-password"
			hint={text.newPasswordHint}
			bind:value={next}
			error={editor.errors.newPassword}
		/>
		<div class="form-footer">
			<Button variant="secondary" onclick={close}>{text.cancel}</Button>
			<Button type="submit" busy={editor.busy}>{text.savePassword}</Button>
		</div>
	</form>
</Dialog>
