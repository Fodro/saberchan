<script lang="ts">
	import { t } from '$lib/translations';
	import * as Card from '$lib/components/ui/card/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import { invalidate } from '$app/navigation';
	import PostsChart from './PostsChart.svelte';

	let { data } = $props();

	let dailyMetrics = $derived(data.dailyMetrics);
	let error = $derived<string | null>(data.error ?? null);
	let loading = $state(false);

	async function refresh() {
		loading = true;
		try {
			await invalidate('admin:metrics');
		} catch (e) {
			error = String(e);
		} finally {
			loading = false;
		}
	}
</script>

<div class="container mx-auto p-6 space-y-6">
	<Card.Root>
		<Card.Header>
			<Card.Title class="flex flex-row items-center justify-between">
				<h2>{$t("common.admin.dashboard.title")}</h2>
				<Button variant="outline" size="sm" onclick={refresh} disabled={loading}>
					{$t("common.admin.dashboard.refresh")}
				</Button>
			</Card.Title>
			<Card.Description>
				{$t("common.admin.dashboard.description")}
			</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if error}
				<div class="text-destructive p-4 bg-destructive/10 rounded">
					{$t("common.admin.dashboard.error")}{error}
				</div>
			{:else if dailyMetrics === null || dailyMetrics === undefined}
				<div class="text-muted-foreground p-4">
					{$t("common.admin.dashboard.noData")}
				</div>
			{:else}
				<div class="space-y-6">
					<PostsChart {dailyMetrics} />
					<details class="text-sm text-muted-foreground">
						<summary class="cursor-pointer mb-2">{$t("common.admin.dashboard.rawData")}</summary>
						<div class="overflow-auto">
							<pre class="font-mono bg-muted p-4 rounded max-h-[40vh] overflow-auto whitespace-pre-wrap">{JSON.stringify(dailyMetrics, null, 2)}</pre>
						</div>
					</details>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
</div>