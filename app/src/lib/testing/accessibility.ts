import axe from 'axe-core';
import { expect } from 'vitest';
import '#lib/styles/app.css';

const pageLevelRules = ['region', 'landmark-one-main', 'page-has-heading-one'];

export async function expectAccessible(container: Element): Promise<void> {
	const rules = Object.fromEntries(pageLevelRules.map((rule) => [rule, { enabled: false }]));
	const results = await axe.run({ include: [container] }, { resultTypes: ['violations'], rules });
	const summary = results.violations.map(
		(violation) =>
			`${violation.id}: ${violation.help} (${violation.nodes.map((node) => `${node.target.join(', ')} ${node.failureSummary ?? ''}`).join('; ')})`
	);
	expect(summary).toEqual([]);
}
