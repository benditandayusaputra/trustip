<script lang="ts">
	import LegalPage from '#lib/features/legal/components/LegalPage.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<LegalPage
	document={id.legal.privacy}
	related={{ href: '/ketentuan', title: id.legal.terms.title }}
/>
