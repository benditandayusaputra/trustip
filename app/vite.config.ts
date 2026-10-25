import tailwindcss from '@tailwindcss/vite';
import { loadEnv } from 'vite';
import { defineConfig } from 'vitest/config';
import type { TestProjectInlineConfiguration } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import adapter from '@sveltejs/adapter-vercel';
import { sveltekit } from '@sveltejs/kit/vite';
import { contentSecurityPolicy } from './csp.config';

const versionPollMs = 300_000;
const vercelRegion = 'sin1';

const isLibraryFile = (filename: string) => filename.split(/[/\\]/).includes('node_modules');

function kitPlugin(mode: string) {
	const publicEnv = loadEnv(mode, process.cwd(), 'PUBLIC_');
	return sveltekit({
		compilerOptions: {
			runes: ({ filename }) => (isLibraryFile(filename) ? undefined : true)
		},
		adapter: adapter({ regions: [vercelRegion] }),
		version: { pollInterval: versionPollMs },
		prerender: { crawl: false },
		csp: {
			mode: 'hash',
			directives: contentSecurityPolicy({
				files: publicEnv.PUBLIC_FILES_BASE_URL,
				stream: publicEnv.PUBLIC_STREAM_URL
			})
		}
	});
}

const browserTests: TestProjectInlineConfiguration = {
	extends: './vite.config.ts',
	test: {
		expect: { requireAssertions: true },
		name: 'client',
		browser: {
			enabled: true,
			provider: playwright(),
			instances: [{ browser: 'chromium', headless: true }]
		},
		include: ['src/**/*.svelte.test.ts'],
		exclude: ['src/lib/server/**']
	}
};

const nodeTests: TestProjectInlineConfiguration = {
	extends: './vite.config.ts',
	test: {
		expect: { requireAssertions: true },
		name: 'server',
		environment: 'node',
		include: ['src/**/*.test.ts', '*.test.ts'],
		exclude: ['src/**/*.svelte.test.ts']
	}
};

const toolingTests: TestProjectInlineConfiguration = {
	test: {
		name: 'tooling',
		environment: 'node',
		include: ['eslint/**/*.test.js', 'scripts/**/*.test.js']
	}
};

export default defineConfig(({ mode }) => ({
	plugins: [tailwindcss(), kitPlugin(mode)],
	optimizeDeps: {
		include: ['@lucide/svelte/icons/*']
	},
	test: {
		passWithNoTests: true,
		projects: [browserTests, nodeTests, toolingTests]
	}
}));
