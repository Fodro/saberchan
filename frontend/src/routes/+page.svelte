<script lang="ts">
	import { HoverCard, HoverCardTrigger } from "$lib/components/ui/hover-card";
	import { Separator } from "$lib/components/ui/separator/index.js";
	import HoverCardContent from "$lib/components/ui/hover-card/hover-card-content.svelte";
	import { t } from "$lib/translations";

	let { data } = $props();
</script>

<svelte:head>
	<title>Saberchan - {$t("common.boards.list_title")}</title>
</svelte:head>

<h3 class="mt-8 scroll-m-20 text-2xl font-semibold tracking-tight mb-5">
	{$t("common.boards.list_title")}
</h3>

<Separator class="my-4" />

<div class="flex flex-col gap-3 justify-start items-start h-auto">
	{#each data.boards as board}
		<div class="flex flex-row justify-start items-center">
			<HoverCard>
				<HoverCardTrigger
					href={`/board/${board.alias}`}
					target="_blank"
					rel="noreferrer noopener"
					class="rounded-sm underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-8 focus-visible:outline-black"
				>
					/{board.alias}/ - {board.name}
				</HoverCardTrigger>
				<HoverCardContent>
					{board.description}
				</HoverCardContent>
			</HoverCard>
		</div>
	{/each}
</div>
