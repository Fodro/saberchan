<script lang="ts">
	import { DrawingPin, DrawingPinFilled } from "svelte-radix";
	import { t } from "$lib/translations";
	import Button from "../ui/button/button.svelte";

	const {
		initialLeft,
		initialTop,
	}: {
		initialLeft: number;
		initialTop: number;
	} = $props();

	let left = $state(initialLeft + 100);
	let top = $state(initialTop + 100);

	let pinned = $state(false);

	let moving = $state(false);

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
<!-- svelte-ignore slot_element_deprecated -->
<!-- svelte-ignore event_directive_deprecated -->
<section
	on:mousedown={onMouseDown}
	style="left: {left}px; top: {top}px;"
	class="draggable w-[50vw] h-[70vh]"
>
	<div class="flex flex-row-reverse items-center h-[5%]">
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
