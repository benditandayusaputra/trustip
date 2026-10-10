<script lang="ts">
	import Minus from '@lucide/svelte/icons/minus';
	import Plus from '@lucide/svelte/icons/plus';
	import FieldShell from './FieldShell.svelte';
	import { describedBy, inputClass } from './field';
	import { common as id } from '#lib/i18n/common.js';
	import { parseDigits } from '#lib/i18n/format.js';

	interface Props {
		label: string;
		value?: number;
		min?: number;
		max?: number;
		name?: string;
		hint?: string | undefined;
		error?: string | undefined;
	}

	let {
		label,
		value = $bindable(0),
		min = 0,
		max = Number.MAX_SAFE_INTEGER,
		name,
		hint,
		error
	}: Props = $props();

	const inputId = $props.id();
	const clamp = (next: number) => Math.min(max, Math.max(min, next));
	const stepperClass =
		'inline-flex size-12 shrink-0 items-center justify-center rounded-control border ' +
		'border-border-strong text-text disabled:opacity-60';

	function readInput(event: Event & { currentTarget: HTMLInputElement }) {
		const input = event.currentTarget;
		value = clamp(parseDigits(input.value) ?? min);
		if (input.value !== String(value)) {
			input.value = String(value);
		}
	}
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<div class="flex items-center gap-2">
		<button
			type="button"
			class={stepperClass}
			aria-label={id.form.decrease(label)}
			disabled={value <= min}
			onclick={() => (value = clamp(value - 1))}
		>
			<Minus class="size-6" aria-hidden="true" />
		</button>
		<input
			id={inputId}
			{name}
			type="text"
			inputmode="numeric"
			autocomplete="off"
			value={String(value)}
			oninput={readInput}
			aria-invalid={error ? true : undefined}
			aria-describedby={describedBy(inputId, hint, error)}
			class="{inputClass} text-right font-mono font-medium tabular-nums"
		/>
		<button
			type="button"
			class={stepperClass}
			aria-label={id.form.increase(label)}
			disabled={value >= max}
			onclick={() => (value = clamp(value + 1))}
		>
			<Plus class="size-6" aria-hidden="true" />
		</button>
	</div>
</FieldShell>
