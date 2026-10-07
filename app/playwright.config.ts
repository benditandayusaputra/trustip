import { defineConfig, devices } from '@playwright/test';

const appOriginPort = 5173;
const apiHealthUrl = 'http://127.0.0.1:8080/healthz';

export default defineConfig({
	testDir: 'tests',
	forbidOnly: true,
	webServer: [
		{
			command: 'cd ../api && set -a && . ./.env && set +a && go run ./cmd/server',
			url: apiHealthUrl,
			reuseExistingServer: false,
			timeout: 120_000
		},
		{
			command: `npm run build && npm run preview -- --port ${String(appOriginPort)} --strictPort`,
			port: appOriginPort
		}
	],
	use: { baseURL: `http://localhost:${String(appOriginPort)}` },
	projects: [
		{ name: 'mobile', use: { ...devices['Pixel 7'] } },
		{ name: 'desktop', use: { ...devices['Desktop Chrome'] } }
	]
});
