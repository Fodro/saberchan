<script lang="ts">
	import MarkupNodes from "$lib/components/custom/MarkupNodes.svelte";
	import { buildPostLines } from "$lib/markup";

	let { text, additionalClass }: { text: string; additionalClass: string } = $props();

	const lines = $derived(buildPostLines(text));
	const textSize = $derived(text.length > 1000 ? "text-sm" : "text-base");
</script>

<div class="flex-2">
	{#each lines as line, i (i)}
		{#if line.kind === "reply"}
			<a
				class={`text-orange-500 hover:underline cursor-pointer ${textSize} ${additionalClass}`}
				href={line.replyNumber ? `#${line.replyNumber}` : undefined}
			>
				<MarkupNodes nodes={line.nodes} />
			</a>
		{:else if line.kind === "greentext"}
			<p class={`text-green-500 ${textSize} ${additionalClass}`}>
				<MarkupNodes nodes={line.nodes} />
			</p>
		{:else}
			<p class={`${textSize} break-normal ${additionalClass}`}>
				<MarkupNodes nodes={line.nodes} />
			</p>
		{/if}
	{/each}
</div>
