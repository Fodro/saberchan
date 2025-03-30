<script lang="ts">
	import { DrawingPin, DrawingPinFilled } from "svelte-radix";
	import { t } from "$lib/translations";
	import Button from "../ui/button/button.svelte";

	export let left = 100;
	export let top = 100;

	let pinned = false;

	let moving = false;

	function onMouseDown() {
		if (pinned) {
			return;
		}
		moving = true;
	}

	function onMouseMove(e: any) {
		if (moving) {
			left += e.movementX;
			top += e.movementY;
		}
	}

	function onMouseUp() {
		moving = false;
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<section
	on:mousedown={onMouseDown}
	style="left: {left}px; top: {top}px;"
	class="draggable w-auto h-auto"
>
	<div class="flex flex-row-reverse items-center">
		<Button
			class="cursor-pointer"
			variant="ghost"
			size="icon"
			on:click={() => {
				pinned = !pinned;
			}}
		>
			{#if pinned}
				<DrawingPinFilled />
			{/if}
			{#if !pinned}
				<DrawingPin />
			{/if}
		</Button>
		<p class="text-muted-foreground">{$t("common.draggable")}</p>
	</div>
	<slot></slot>
</section>

<svelte:window on:mouseup={onMouseUp} on:mousemove={onMouseMove} />

<style>
	.draggable {
		user-select: none;
		cursor: move;
		border: solid 1px gray;
		position: absolute;
	}
</style>
