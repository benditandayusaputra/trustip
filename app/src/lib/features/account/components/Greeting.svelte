<script lang="ts">
	import { id } from '#lib/i18n/id.js';
	import { getAccountSession } from '../session.svelte';

	const session = getAccountSession();
</script>

<h1 class="text-xl font-semibold text-text">
	{id.account.greeting(session.account?.full_name ?? '')}
</h1>
