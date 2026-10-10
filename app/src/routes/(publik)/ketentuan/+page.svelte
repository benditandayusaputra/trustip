<script lang="ts">
	import LegalPage from '#lib/features/legal/components/LegalPage.svelte';
	import { legal as legalText } from '#lib/i18n/legal.js';

	const id = { legal: legalText };
</script>

<LegalPage
	document={id.legal.terms}
	related={{ href: '/privasi', title: id.legal.privacy.title }}
/>
