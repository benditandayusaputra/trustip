<script lang="ts">
	import { apiClient } from '#lib/api/client.js';
	import { getMeta } from '#lib/api/meta.js';
	import { common as id } from '#lib/i18n/common.js';

	let version = $state('');

	$effect(() => {
		getMeta(apiClient())
			.then((meta) => {
				version = meta.version;
			})
			.catch(() => {
				version = '';
			});
	});
</script>

{#if version}
	<p data-testid="api-version">{id.system.apiVersion(version)}</p>
{/if}
