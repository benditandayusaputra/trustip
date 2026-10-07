<script lang="ts">
	import { badgeTone, type Tone } from './tone';

	interface Props {
		tone: Tone;
		label: string;
	}

	let { tone, label }: Props = $props();
</script>

<span
	class="inline-flex items-center rounded-pill px-2 py-1 text-xs leading-tight font-medium {badgeTone[
		tone
	]}">{label}</span
>
