<script lang="ts">
	import { onMount } from 'svelte';
	import {
		loadTurnstile,
		turnstileAppearance,
		type TurnstileApi,
		type TurnstileAppearance
	} from './turnstile';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		siteKey: string;
		action: string;
		onchange: (token: string) => void;
		error?: string | undefined;
	}

	let { siteKey, action, onchange, error }: Props = $props();

	const labelId = $props.id();
	let container: HTMLDivElement;
	let widget: ({ api: TurnstileApi; widgetId: string } & TurnstileAppearance) | undefined;
	let size = $state<TurnstileAppearance['size']>('compact');
	let loadFailed = $state(false);

	function mount(api: TurnstileApi): void {
		const appearance = turnstileAppearance(container);
		if (widget?.theme === appearance.theme && widget.size === appearance.size) {
			return;
		}
		widget?.api.remove(widget.widgetId);
		onchange('');
		size = appearance.size;
		const widgetId = api.render(container, {
			sitekey: siteKey,
			action,
			...appearance,
			language: 'id',
			callback: onchange,
			'expired-callback': () => {
				onchange('');
			},
			'error-callback': () => {
				onchange('');
			}
		});
		widget = { api, widgetId, ...appearance };
	}

	function refresh(): void {
		if (widget) {
			mount(widget.api);
		}
	}

	function watchAppearance(): () => void {
		let frame = 0;
		const resize = new ResizeObserver(() => {
			cancelAnimationFrame(frame);
			frame = requestAnimationFrame(refresh);
		});
		resize.observe(container);
		window.addEventListener('trustip:appearance', refresh);
		return () => {
			resize.disconnect();
			cancelAnimationFrame(frame);
			window.removeEventListener('trustip:appearance', refresh);
		};
	}

	onMount(() => {
		let isMounted = true;
		const stopWatching = watchAppearance();
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
			stopWatching();
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
	<div
		bind:this={container}
		class={size === 'compact' ? 'min-h-(--captcha-compact-height)' : 'min-h-(--captcha-height)'}
	></div>
	{#if loadFailed}
		<p role="alert" class="text-sm text-danger">{id.captcha.loadFailed}</p>
	{/if}
	{#if error}
		<p class="text-sm text-danger">{error}</p>
	{/if}
</div>
