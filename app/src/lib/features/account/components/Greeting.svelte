<script lang="ts">
	import { common as id } from '#lib/i18n/common.js';
	import { getAccountSession } from '../session.svelte';

	const session = getAccountSession();
</script>

<h1 class="text-xl font-semibold text-text">
	{id.account.greeting(session.account?.full_name ?? '')}
</h1>
