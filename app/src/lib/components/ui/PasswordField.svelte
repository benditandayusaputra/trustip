<script lang="ts">
	import Eye from '@lucide/svelte/icons/eye';
	import EyeOff from '@lucide/svelte/icons/eye-off';
	import FieldShell from './FieldShell.svelte';
	import { describedBy, inputClass } from './field';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		label: string;
		value?: string;
		name?: string;
		autocomplete?: 'current-password' | 'new-password';
		hint?: string | undefined;
		error?: string | undefined;
	}

	let { label, value = $bindable(''), name, autocomplete, hint, error }: Props = $props();
	let visible = $state(false);
	const inputId = $props.id();
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<div class="relative">
		<input
			id={inputId}
			{name}
			type={visible ? 'text' : 'password'}
			{autocomplete}
			bind:value
			aria-invalid={error ? true : undefined}
			aria-describedby={describedBy(inputId, hint, error)}
			class="{inputClass} pr-12"
		/>
		<button
			type="button"
			aria-label={visible ? id.auth.hidePassword : id.auth.showPassword}
			aria-pressed={visible}
			aria-controls={inputId}
			onclick={() => (visible = !visible)}
			class="absolute inset-y-0 right-0 flex min-w-11 items-center justify-center rounded-control text-text-muted hover:text-text"
		>
			{#if visible}<EyeOff class="size-5" aria-hidden="true" />{:else}<Eye
					class="size-5"
					aria-hidden="true"
				/>{/if}
		</button>
	</div>
</FieldShell>
