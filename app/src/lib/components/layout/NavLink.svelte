<script lang="ts">
	import { isCurrent, type NavItem } from './navigation';

	interface Props {
		item: NavItem;
		currentPath: string;
		layout: 'tab' | 'row';
	}

	let { item, currentPath, layout }: Props = $props();

	const current = $derived(isCurrent(item.href, currentPath));
	const layouts = {
		tab: 'flex min-h-11 flex-1 flex-col items-center justify-center gap-1 text-xs',
		row: 'flex min-h-11 items-center gap-4 rounded-control px-4 text-base'
	};
</script>

<a
	href={item.href}
	aria-current={current ? 'page' : undefined}
	class="{layouts[layout]} {current
		? 'font-semibold text-accent'
		: 'text-text-muted hover:text-text'}"
>
	<item.icon class="size-6 shrink-0" aria-hidden="true" />
	<span>{item.label}</span>
</a>
