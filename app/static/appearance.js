const appearanceStorageKey = 'trustip-appearance';
const appearanceChoices = new Set(['light', 'dark']);
const defaultAppearance = 'light';
const appearanceRoot = document.documentElement;

function readAppearance() {
	try {
		return localStorage.getItem(appearanceStorageKey) ?? defaultAppearance;
	} catch {
		return defaultAppearance;
	}
}

function applyAppearance(preference) {
	const chosen = appearanceChoices.has(preference) ? preference : defaultAppearance;
	appearanceRoot.dataset.preference = chosen;
	appearanceRoot.dataset.appearance = chosen;
	syncAppearanceControls(chosen);
	const themeMeta = document.querySelector('meta[name="theme-color"]');
	themeMeta?.setAttribute(
		'content',
		getComputedStyle(appearanceRoot).getPropertyValue('--color-bg')
	);
	window.dispatchEvent(new Event('trustip:appearance'));
}

function syncAppearanceControls(chosen) {
	for (const toggle of document.querySelectorAll('[data-appearance-toggle]')) {
		if (toggle instanceof HTMLInputElement) {
			toggle.checked = chosen === 'dark';
		}
	}
	for (const choice of document.querySelectorAll('[data-appearance-choice]')) {
		choice.setAttribute(
			'aria-pressed',
			String(choice.getAttribute('data-appearance-choice') === chosen)
		);
	}
}

function closeAppearanceMenus(keep) {
	for (const menu of document.querySelectorAll('details[data-appearance-menu][open]')) {
		if (menu !== keep) {
			menu.removeAttribute('open');
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
document.addEventListener('change', (event) => {
	const toggle = event.target;
	if (toggle instanceof HTMLInputElement && toggle.hasAttribute('data-appearance-toggle')) {
		saveAppearance(toggle.checked ? 'dark' : 'light');
	}
});
document.addEventListener('click', (event) => {
	const target = event.target instanceof Element ? event.target : null;
	const choice = target?.closest('[data-appearance-choice]');
	if (choice) {
		saveAppearance(choice.getAttribute('data-appearance-choice'));
		closeAppearanceMenus(null);
		return;
	}
	closeAppearanceMenus(target?.closest('details[data-appearance-menu]') ?? null);
});
document.addEventListener('keydown', (event) => {
	const open = document.querySelector('details[data-appearance-menu][open]');
	if (event.key === 'Escape' && open) {
		closeAppearanceMenus(null);
		open.querySelector('summary')?.focus();
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
