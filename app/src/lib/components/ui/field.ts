export function describedBy(inputId: string, hint?: string, error?: string): string | undefined {
	const ids = [hint ? `${inputId}-hint` : '', error ? `${inputId}-error` : ''].filter(Boolean);
	return ids.length > 0 ? ids.join(' ') : undefined;
}

export const inputClass =
	'field-control min-h-12 w-full rounded-control border border-border-strong bg-surface px-4 text-base text-text ' +
	'focus:border-accent aria-invalid:border-danger';
