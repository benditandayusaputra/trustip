<script lang="ts">
	import { isCurrent, type NavItem } from './navigation';

	interface Props {
		item: NavItem;
		currentPath: string;
		layout: 'tab' | 'row';
		onclick?: () => void;
	}

	let { item, currentPath, layout, onclick }: Props = $props();

	const current = $derived(isCurrent(item.href, currentPath));
	const layouts = {
		tab: 'flex min-h-11 flex-1 flex-col items-center justify-center gap-1 text-xs',
		row: 'flex min-h-11 items-center gap-2 rounded-control px-4 text-sm'
	};
</script>

<a
	href={item.href}
	{onclick}
	aria-current={current ? 'page' : undefined}
	class="nav-link {layouts[layout]}"
>
	<item.icon class="size-5 shrink-0" aria-hidden="true" />
	<span>{item.label}</span>
</a>
