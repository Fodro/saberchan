<script lang="ts">
	import MarkupNodes from "$lib/components/custom/MarkupNodes.svelte";
	import { MARKUP_TAG_CLASS, type MarkupNode } from "$lib/markup";

	let { nodes }: { nodes: MarkupNode[] } = $props();
</script>

{#each nodes as node, i (i)}
	{#if node.type === "text"}
		{node.value}
	{:else}
		{#if node.tag === "spoiler"}
			<button class={MARKUP_TAG_CLASS[node.tag]} type="button">
				<MarkupNodes nodes={node.children} />
			</button>
		{:else}
			<span class={MARKUP_TAG_CLASS[node.tag]}>
				<MarkupNodes nodes={node.children} />
			</span>
		{/if}
	{/if}
{/each}
