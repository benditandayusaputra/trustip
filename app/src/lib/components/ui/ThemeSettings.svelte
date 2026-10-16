<script lang="ts">
	import Card from './Card.svelte';
	import ThemePicker from './ThemePicker.svelte';
	import { appearance as text } from '#lib/i18n/appearance.js';
</script>

<Card title={text.title}>
	<p class="text-sm text-text-muted">{text.intro}</p>
	<div><ThemePicker expanded /></div>
</Card>
