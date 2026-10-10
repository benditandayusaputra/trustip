import { defineConfig, devices } from '@playwright/test';
import base from './playwright.config';

export default defineConfig({
	...base,
	workers: 1,
	testMatch: /(?:appearance|responsive|home|security-headers)\.test\.ts/u,
	projects: [
		{
			name: 'firefox',
			use: {
				...devices['Desktop Firefox'],
				colorScheme: 'dark',
				launchOptions: { firefoxUserPrefs: { 'ui.systemUsesDarkTheme': 1 } }
			}
		},
		{ name: 'webkit-desktop', use: { ...devices['Desktop Safari'] } },
		{ name: 'webkit-mobile', use: { ...devices['iPhone 13'] } }
	]
});
