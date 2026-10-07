<script lang="ts">
	import { onMount } from 'svelte';
	import { loadTurnstile, type TurnstileApi } from './turnstile';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		siteKey: string;
		action: string;
		onchange: (token: string) => void;
		error?: string | undefined;
	}

	let { siteKey, action, onchange, error }: Props = $props();

	const labelId = $props.id();
	let container: HTMLDivElement;
	let widget: { api: TurnstileApi; widgetId: string } | undefined;
	let loadFailed = $state(false);

	function mount(api: TurnstileApi): void {
		const widgetId = api.render(container, {
			sitekey: siteKey,
			action,
			theme: 'dark',
			language: 'id',
			callback: (solved) => {
				onchange(solved);
			},
			'expired-callback': () => {
				onchange('');
			},
			'error-callback': () => {
				onchange('');
			}
		});
		widget = { api, widgetId };
	}

	onMount(() => {
		let isMounted = true;
		loadTurnstile()
			.then((api) => {
				if (isMounted) {
					mount(api);
				}
			})
			.catch(() => {
				loadFailed = true;
			});
		return () => {
			isMounted = false;
			widget?.api.remove(widget.widgetId);
		};
	});

	export function reset(): void {
		onchange('');
		widget?.api.reset(widget.widgetId);
	}
</script>

<div role="group" aria-labelledby={labelId} class="flex flex-col gap-2">
	<p id={labelId} class="text-sm font-medium text-text">{id.captcha.label}</p>
	<div bind:this={container} class="min-h-(--captcha-height)"></div>
	{#if loadFailed}
		<p role="alert" class="text-sm text-danger">{id.captcha.loadFailed}</p>
	{/if}
	{#if error}
		<p class="text-sm text-danger">{error}</p>
	{/if}
</div>
