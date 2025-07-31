<script lang="ts">
	import Button from "../ui/button/button.svelte";
	import * as Card from "$lib/components/ui/card/index.js";

	let {
		link,
		name,
	}: {
		link: string;
		name: string;
	} = $props();

	let viewFull: boolean = $state(false);
	
</script>

<div class="flex flex-col justify-center items-center">
	<Button
		variant="link"
		class="cursor-pointer"
		on:click={() => {
			viewFull = !viewFull;
		}}
	>
		{name}
	</Button>
	{#if viewFull}
		<img src={link} alt={name} class="h-100 w-120 object-scale-down" />
	{:else}
		<img src={link} alt={name} class="h-48 w-72 object-scale-down" />
	{/if}
</div>
