import { readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';

const tokenFiles = new Set(['src/lib/styles/tokens.css', 'src/lib/styles/tokens.ts']);
const ignoredFiles = new Set(['src/lib/api/schema.d.ts']);
const scannedExtensions = new Set(['.svelte', '.ts', '.css', '.html']);
const allowedSpacingSteps = new Set(['0', '1', '2', '4', '5', '6', '8', '11', '12']);

const palette =
	'slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|' +
	'blue|indigo|violet|purple|fuchsia|pink|rose|white|black';
const colorUtilities =
	'bg|text|border|ring|fill|stroke|from|via|to|outline|decoration|accent|caret|divide|shadow';
const spacingUtilities =
	'p|px|py|pt|pr|pb|pl|ps|pe|m|mx|my|mt|mr|mb|ml|ms|me|gap|gap-x|gap-y|space-x|space-y|' +
	'inset|inset-x|inset-y|top|right|bottom|left|w|h|min-w|min-h|max-w|max-h|size';

const rules = [
	{
		name: 'raw-hex-color',
		pattern: /(?<![\w&])#(?:[0-9a-fA-F]{8}|[0-9a-fA-F]{6}|[0-9a-fA-F]{3,4})\b/gu
	},
	{
		name: 'raw-color-function',
		pattern: /\b(?:rgba?|hsla?|hwb|lab|lch|oklab|oklch)\(/gu
	},
	{
		name: 'default-palette-class',
		pattern: new RegExp(`\\b(?:${colorUtilities})-(?:${palette})(?:-\\d{2,3})?\\b`, 'gu')
	},
	{
		name: 'arbitrary-size-class',
		pattern: /\b[a-z][\w-]*-\[[^\]]*\d[^\]]*\]/gu
	},
	{
		name: 'raw-css-size',
		pattern: /(?<![\w-])(?:font-size|margin|padding|gap)[a-z-]*\s*:\s*-?\d/gu
	}
];

const spacingClass = new RegExp(
	`(?<![\\w-])-?(?:${spacingUtilities})-(\\d+(?:\\.\\d+)?)(?![\\w.-])`,
	'gu'
);

function lineOf(text, index) {
	return text.slice(0, index).split('\n').length;
}

function spacingViolations(text) {
	return [...text.matchAll(spacingClass)]
		.filter((match) => !allowedSpacingSteps.has(match[1] ?? ''))
		.map((match) => ({ line: lineOf(text, match.index), rule: 'spacing-outside-tokens' }));
}

export function findViolations(relativePath, text) {
	if (tokenFiles.has(relativePath) || ignoredFiles.has(relativePath)) {
		return [];
	}
	const ruleViolations = rules.flatMap((rule) =>
		[...text.matchAll(rule.pattern)].map((match) => ({
			line: lineOf(text, match.index),
			rule: rule.name
		}))
	);
	return [...ruleViolations, ...spacingViolations(text)];
}

function listFiles(directory) {
	return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
		const fullPath = path.join(directory, entry.name);
		return entry.isDirectory() ? listFiles(fullPath) : [fullPath];
	});
}

function run() {
	const violations = listFiles('src')
		.filter((file) => scannedExtensions.has(path.extname(file)))
		.flatMap((file) =>
			findViolations(file, readFileSync(file, 'utf8')).map((found) => ({ file, ...found }))
		);
	violations.forEach((found) => {
		process.stderr.write(`${found.file}:${String(found.line)}: ${found.rule}\n`);
	});
	process.exitCode = violations.length > 0 ? 1 : 0;
}

if (import.meta.url === `file://${process.argv[1] ?? ''}`) {
	run();
}
