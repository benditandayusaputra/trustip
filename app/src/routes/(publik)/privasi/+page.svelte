<script lang="ts">
	import LegalPage from '#lib/features/legal/components/LegalPage.svelte';
	import { legal as legalText } from '#lib/i18n/legal.js';

	const id = { legal: legalText };
</script>

<LegalPage
	document={id.legal.privacy}
	related={{ href: '/ketentuan', title: id.legal.terms.title }}
/>
