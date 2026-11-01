import { ApiError } from '#lib/api/errors.js';
import type { Organization } from '#lib/api/organizations.js';
import type { Toaster } from '#lib/components/ui/toast.svelte.js';

export type SectionOutcome =
	| { kind: 'saved'; organization: Organization }
	| { kind: 'invalid' }
	| { kind: 'failed'; message: string };

export interface SectionFailure<F extends string> {
	outcome: SectionOutcome;
	errors: Partial<Record<F, string>>;
}

export function explainSectionFailure<F extends string>(
	error: unknown,
	fieldsFromServer: Record<string, F | undefined>
): SectionFailure<F> {
	if (!(error instanceof ApiError) || error.fields.length === 0) {
		const message = error instanceof Error ? error.message : '';
		return { outcome: { kind: 'failed', message }, errors: {} };
	}
	const errors: Partial<Record<F, string>> = {};
	error.fields.forEach((problem) => {
		const field = fieldsFromServer[problem.field];
		if (field) {
			errors[field] = problem.message;
		}
	});
	return { outcome: { kind: 'invalid' }, errors };
}

export interface SectionReport {
	toaster: Toaster;
	savedMessage?: string;
	onsaved: (organization: Organization) => void;
}

export function reportSection(outcome: SectionOutcome, report: SectionReport): void {
	if (outcome.kind === 'saved') {
		if (report.savedMessage) {
			report.toaster.show({ tone: 'success', message: report.savedMessage });
		}
		report.onsaved(outcome.organization);
	} else if (outcome.kind === 'failed') {
		report.toaster.show({ tone: 'danger', message: outcome.message });
	}
}
