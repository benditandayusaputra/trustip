import { RuleTester } from 'eslint';
import svelteParser from 'svelte-eslint-parser';
import path from 'node:path';
import { describe, it } from 'vitest';
import { noComments } from './no-comments.js';
import { layerImports } from './layer-imports.js';

RuleTester.describe = describe;
RuleTester.it = it;
RuleTester.itOnly = it.only;

const tester = new RuleTester({
	linterOptions: { noInlineConfig: true },
	languageOptions: { ecmaVersion: 'latest', sourceType: 'module' }
});
const svelteTester = new RuleTester({ languageOptions: { parser: svelteParser } });

const inSource = (relative) => path.join(process.cwd(), 'src', relative);

tester.run('no-comments', noComments, {
	valid: [
		{ code: "const url = 'https://trustip.id/a';" },
		{ code: '#!/usr/bin/env node\nconst ready = true;' }
	],
	invalid: [
		{ code: '// note\nconst ready = true;', errors: [{ messageId: 'comment' }] },
		{ code: 'const ready = /* note */ true;', errors: [{ messageId: 'comment' }] },
		{
			code: '/* eslint-disable */\nconst ready = true;',
			errors: [
				{ message: /has no effect because you have 'noInlineConfig'/u },
				{ messageId: 'comment' }
			]
		}
	]
});

svelteTester.run('no-comments in svelte markup', noComments, {
	valid: [{ code: '<p>a -- b</p>' }],
	invalid: [
		{ code: '<!-- note -->\n<p>a</p>', errors: [{ messageId: 'comment' }] },
		{ code: '<script>\n// note\n</script>', errors: [{ messageId: 'comment' }] }
	]
});

tester.run('layer-imports', layerImports, {
	valid: [
		{ code: "import { x } from '#lib/api/client';", filename: inSource('routes/+page.ts') },
		{
			code: "import { x } from '#lib/components/ui/Button.svelte';",
			filename: inSource('lib/features/drop/a.ts')
		},
		{ code: "import { x } from './state.svelte';", filename: inSource('lib/features/drop/a.ts') },
		{ code: "import { x } from './money';", filename: inSource('lib/domain/fefo.ts') },
		{ code: "import { x } from '#lib/domain/money';", filename: inSource('lib/offline/outbox.ts') },
		{
			code: "import { x } from '#lib/testing/accessibility.js';",
			filename: inSource('lib/components/ui/controls.svelte.test.ts')
		}
	],
	invalid: [
		{
			code: "import { x } from '#lib/features/stock/state.svelte';",
			filename: inSource('lib/features/drop/a.ts'),
			errors: [{ messageId: 'forbidden' }]
		},
		{
			code: "import { x } from '#lib/api/client';",
			filename: inSource('lib/components/ui/Button.ts'),
			errors: [{ messageId: 'forbidden' }]
		},
		{
			code: "import { x } from '../i18n/id';",
			filename: inSource('lib/domain/money.ts'),
			errors: [{ messageId: 'forbidden' }]
		},
		{
			code: "import { x } from '#lib/domain/fefo';",
			filename: inSource('routes/+page.ts'),
			errors: [{ messageId: 'forbidden' }]
		},
		{
			code: "import { x } from '#lib/testing/accessibility.js';",
			filename: inSource('lib/components/ui/Button.ts'),
			errors: [{ messageId: 'forbidden' }]
		},
		{
			code: "import { x } from '$lib/features/drop/a';",
			filename: inSource('lib/offline/sync.ts'),
			errors: [{ messageId: 'forbidden' }]
		}
	]
});
