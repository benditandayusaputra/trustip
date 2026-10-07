import { defineConfig, devices } from '@playwright/test';

const previewPort = 4173;

export default defineConfig({
	testDir: 'tests',
	forbidOnly: true,
	webServer: { command: 'npm run build && npm run preview', port: previewPort },
	use: { baseURL: `http://localhost:${String(previewPort)}` },
	projects: [
		{ name: 'mobile', use: { ...devices['Pixel 7'] } },
		{ name: 'desktop', use: { ...devices['Desktop Chrome'] } }
	]
});
