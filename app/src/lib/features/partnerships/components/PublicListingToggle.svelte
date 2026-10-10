<script lang="ts">
	import Switch from '#lib/components/ui/Switch.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Partnership } from '#lib/api/partnerships.js';
	import { partnerships as id } from '#lib/i18n/partnerships.js';
	import { EndingRecord } from '../ending.svelte';

	interface Props {
		partnership: Partnership;
		onchanged: (partnership: Partnership) => void;
	}

	let { partnership, onchanged }: Props = $props();

	const text = id.partners;
	const toaster = getToaster();
	const record = new EndingRecord(apiClient);

	async function change(showOnPublic: boolean) {
		const outcome = await record.setListing(partnership.id, showOnPublic);
		if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
			return;
		}
		toaster.show({ tone: 'success', message: text.listingSaved });
		onchanged(outcome.partnership);
	}
</script>

<Switch
	label={text.showOnPublic}
	description={text.showOnPublicHint}
	checked={partnership.show_on_public}
	busy={record.busy}
	onchange={(checked: boolean) => void change(checked)}
/>
