<script lang="ts">
	import ClipboardCheck from '@lucide/svelte/icons/clipboard-check';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import Card from '#lib/components/ui/Card.svelte';
	import { dashboard as id } from '#lib/i18n/dashboard.js';

	export interface TaskLink {
		label: string;
		href: string;
	}

	interface Props {
		tasks: TaskLink[];
	}

	let { tasks }: Props = $props();
</script>

<Card icon={ClipboardCheck} fill title={id.dashboard.tasksTitle}>
	{#if tasks.length === 0}
		<p class="text-sm text-text-muted">{id.dashboard.noTasks}</p>
	{:else}
		<ul class="flex flex-col divide-y divide-border">
			{#each tasks as task (task.href + task.label)}
				<li>
					<a
						href={task.href}
						class="flex min-h-12 items-center justify-between gap-4 py-2 text-base text-text"
					>
						{task.label}
						<ChevronRight class="size-5 shrink-0 text-text-muted" aria-hidden="true" />
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</Card>
