import prettier from 'eslint-config-prettier';
import path from 'node:path';
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import { defineConfig, includeIgnoreFile } from 'eslint/config';
import globals from 'globals';
import ts from 'typescript-eslint';
import { noComments } from './eslint/no-comments.js';
import { layerImports } from './eslint/layer-imports.js';

const gitignorePath = path.resolve(import.meta.dirname, '..', '.gitignore');

const trustip = { rules: { 'no-comments': noComments, 'layer-imports': layerImports } };

const forbiddenSinks = [
	{
		selector: 'MemberExpression[property.name=/^(innerHTML|outerHTML)$/]',
		message: 'Render text through Svelte instead of writing HTML.'
	},
	{
		selector: "CallExpression[callee.property.name='insertAdjacentHTML']",
		message: 'Render text through Svelte instead of writing HTML.'
	},
	{
		selector: "CallExpression[callee.object.name='document'][callee.property.name='write']",
		message: 'document.write is forbidden.'
	}
];

const limits = {
	functionLines: 30,
	statements: 20,
	params: 3,
	depth: 2,
	complexity: 8,
	fileLines: 250,
	componentLines: 200,
	identifierLength: 2
};

const cleanCodeRules = {
	'max-lines-per-function': ['error', { max: limits.functionLines, skipBlankLines: false }],
	'max-statements': ['error', { max: limits.statements }],
	'max-params': ['error', { max: limits.params }],
	'max-depth': ['error', { max: limits.depth }],
	complexity: ['error', { max: limits.complexity }],
	'max-lines': ['error', { max: limits.fileLines, skipBlankLines: false }],
	'@typescript-eslint/no-magic-numbers': [
		'error',
		{
			ignore: [0, 1, -1],
			ignoreArrayIndexes: true,
			ignoreDefaultValues: true,
			ignoreEnums: true,
			ignoreNumericLiteralTypes: true,
			ignoreReadonlyClassProperties: true,
			ignoreTypeIndexes: true,
			enforceConst: true
		}
	],
	'no-negated-condition': 'error',
	'no-nested-ternary': 'error',
	'no-else-return': ['error', { allowElseIf: false }],
	'no-param-reassign': ['error', { props: true }],
	'id-length': ['error', { min: limits.identifierLength, exceptions: ['i', '_'] }],
	'no-console': 'error'
};

const safetyRules = {
	'no-eval': 'error',
	'no-new-func': 'error',
	'@typescript-eslint/no-implied-eval': 'error',
	'no-restricted-syntax': ['error', ...forbiddenSinks],
	'no-restricted-globals': [
		'error',
		{ name: 'fetch', message: 'Call the API through apiClient.' },
		{ name: 'indexedDB', message: 'Use the offline store.' }
	],
	'svelte/no-at-html-tags': 'error',
	'trustip/no-comments': 'error',
	'trustip/layer-imports': 'error'
};

export default defineConfig(
	includeIgnoreFile(gitignorePath),
	{ ignores: ['.svelte-kit/**', '.vercel/**', 'node_modules/**', 'src/lib/api/schema.d.ts'] },
	{ linterOptions: { noInlineConfig: true, reportUnusedDisableDirectives: 'error' } },
	js.configs.recommended,
	ts.configs.strictTypeChecked,
	ts.configs.stylisticTypeChecked,
	svelte.configs.recommended,
	prettier,
	svelte.configs.prettier,
	{
		plugins: { trustip },
		languageOptions: {
			globals: { ...globals.browser, ...globals.node },
			parserOptions: {
				projectService: true,
				tsconfigRootDir: import.meta.dirname,
				extraFileExtensions: ['.svelte']
			}
		},
		rules: { 'no-undef': 'off', ...cleanCodeRules, ...safetyRules }
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts'],
		languageOptions: { parserOptions: { parser: ts.parser } }
	},
	{
		files: ['**/*.svelte'],
		rules: { 'max-lines': ['error', { max: limits.componentLines, skipBlankLines: false }] }
	},
	{
		files: ['**/*.js'],
		extends: [ts.configs.disableTypeChecked]
	},
	{
		files: ['src/lib/api/client.ts', 'src/lib/server/**'],
		rules: { 'no-restricted-globals': 'off' }
	},
	{
		files: ['**/*.test.ts', '**/*.test.js', 'tests/**'],
		rules: {
			'max-lines-per-function': 'off',
			'max-lines': 'off',
			'max-statements': 'off',
			'@typescript-eslint/no-magic-numbers': 'off'
		}
	}
);
