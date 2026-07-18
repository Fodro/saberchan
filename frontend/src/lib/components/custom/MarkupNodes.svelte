<script lang="ts">
	import MarkupNodes from "$lib/components/custom/MarkupNodes.svelte";
	import { MARKUP_TAG_CLASS, type MarkupNode } from "$lib/markup";

	let { nodes }: { nodes: MarkupNode[] } = $props();
</script>

{#each nodes as node, i (i)}
	{#if node.type === "text"}
		{node.value}
	{:else}
		<span class={MARKUP_TAG_CLASS[node.tag]}>
			<MarkupNodes nodes={node.children} />
		</span>
	{/if}
{/each}
