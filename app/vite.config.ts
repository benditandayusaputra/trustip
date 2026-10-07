import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import adapter from '@sveltejs/adapter-vercel';
import { sveltekit } from '@sveltejs/kit/vite';

const isLibraryFile = (filename: string) => filename.split(/[/\\]/).includes('node_modules');

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				runes: ({ filename }) => (isLibraryFile(filename) ? undefined : true)
			},
			adapter: adapter()
		})
	],
	test: {
		passWithNoTests: true,
		projects: [
			{
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
			},
			{
				extends: './vite.config.ts',
				test: {
					expect: { requireAssertions: true },
					name: 'server',
					environment: 'node',
					include: ['src/**/*.test.ts'],
					exclude: ['src/**/*.svelte.test.ts']
				}
			},
			{
				test: {
					name: 'tooling',
					environment: 'node',
					include: ['eslint/**/*.test.js', 'scripts/**/*.test.js']
				}
			}
		]
	}
});
