<script lang="ts">
	import Users from '@lucide/svelte/icons/users';
	import Boxes from '@lucide/svelte/icons/boxes';
	import ChartNoAxesCombined from '@lucide/svelte/icons/chart-no-axes-combined';
	import TrendingUp from '@lucide/svelte/icons/trending-up';
	import Wallet from '@lucide/svelte/icons/wallet';
	import ShoppingBag from '@lucide/svelte/icons/shopping-bag';
	import HandCoins from '@lucide/svelte/icons/hand-coins';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import type { Tone } from '#lib/components/ui/tone.js';

	const icons = {
		partners: Users,
		stock: Boxes,
		gross: ChartNoAxesCombined,
		profit: TrendingUp,
		payable: Wallet,
		sold: ShoppingBag,
		share: HandCoins
	};
	interface Props {
		metric: keyof typeof icons;
		label: string;
		value: string;
		caption?: string;
		tone?: Tone;
		loading?: boolean;
	}
	let { metric, label, value, caption, tone = 'info', loading = false }: Props = $props();
	const Icon = $derived(icons[metric]);
	const fullOnMobile = $derived(['gross', 'profit', 'payable', 'share'].includes(metric));
</script>

<KpiCard {fullOnMobile} {label} {value} {...caption && { caption }} {tone} {loading}>
	{#snippet icon()}<Icon class="size-5" aria-hidden="true" />{/snippet}
</KpiCard>
