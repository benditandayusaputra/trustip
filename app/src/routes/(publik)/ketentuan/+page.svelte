<script lang="ts">
	import LegalPage from '#lib/features/legal/components/LegalPage.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<LegalPage
	document={id.legal.terms}
	related={{ href: '/privasi', title: id.legal.privacy.title }}
/>
