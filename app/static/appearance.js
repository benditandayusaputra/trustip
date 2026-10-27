const appearanceStorageKey = 'trustip-appearance';
const appearanceChoices = new Set(['system', 'light', 'dark']);
const appearanceRoot = document.documentElement;
const appearanceMedia = window.matchMedia('(prefers-color-scheme: dark)');

function readAppearance() {
	try {
		return localStorage.getItem(appearanceStorageKey) ?? 'system';
	} catch {
		return 'system';
	}
}

function applyAppearance(preference) {
	const chosen = appearanceChoices.has(preference) ? preference : 'system';
	const resolved = appearanceMedia.matches ? 'dark' : 'light';
	appearanceRoot.dataset.preference = chosen;
	appearanceRoot.dataset.appearance = chosen === 'system' ? resolved : chosen;
	syncAppearanceControls(chosen);
	const themeMeta = document.querySelector('meta[name="theme-color"]');
	themeMeta?.setAttribute(
		'content',
		getComputedStyle(appearanceRoot).getPropertyValue('--color-bg')
	);
	window.dispatchEvent(new Event('trustip:appearance'));
}

function syncAppearanceControls(chosen) {
	for (const picker of document.querySelectorAll('[data-appearance-picker]')) {
		if (picker instanceof HTMLSelectElement) {
			picker.value = chosen;
		}
	}
	for (const toggle of document.querySelectorAll('[data-appearance-toggle]')) {
		if (toggle instanceof HTMLInputElement) {
			toggle.checked = appearanceRoot.dataset.appearance === 'dark';
		}
	}
}

function saveAppearance(chosen) {
	if (!appearanceChoices.has(chosen)) {
		return;
	}
	try {
		localStorage.setItem(appearanceStorageKey, chosen);
	} catch {
		appearanceRoot.dataset.preference = chosen;
	}
	applyAppearance(chosen);
}

applyAppearance(readAppearance());
appearanceMedia.addEventListener('change', () =>
	applyAppearance(appearanceRoot.dataset.preference)
);
document.addEventListener('change', (event) => {
	const picker = event.target;
	if (picker instanceof HTMLSelectElement && picker.hasAttribute('data-appearance-picker')) {
		saveAppearance(picker.value);
	}
	if (picker instanceof HTMLInputElement && picker.hasAttribute('data-appearance-toggle')) {
		saveAppearance(picker.checked ? 'dark' : 'light');
	}
});
document.addEventListener('DOMContentLoaded', () =>
	applyAppearance(appearanceRoot.dataset.preference)
);
window.addEventListener('trustip:refresh-appearance', () => {
	applyAppearance(appearanceRoot.dataset.preference);
});
window.addEventListener('storage', (event) => {
	if (event.key === appearanceStorageKey) {
		applyAppearance(readAppearance());
	}
});
