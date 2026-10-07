<script lang="ts">
	import CaptchaField from '#lib/components/ui/CaptchaField.svelte';

	let token = $state('');
	let field: { reset: () => void } | undefined = $state();
</script>

<CaptchaField
	bind:this={field}
	siteKey="site-key"
	action="login"
	onchange={(solved: string) => {
		token = solved;
	}}
/>
<output data-testid="token">{token}</output>
<button type="button" onclick={() => field?.reset()}>Ulangi</button>
