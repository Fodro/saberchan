<script lang="ts">
	import { DrawingPin, DrawingPinFilled } from "svelte-radix";
	import { t } from "$lib/translations";
	import Button from "../ui/button/button.svelte";

	const {
		initialLeft,
		initialTop,
		pinned,
	}: {
		initialLeft: number;
		initialTop: number;
		pinned: boolean;
	} = $props();

	let left = $state(initialLeft + 100);
	let top = $state(initialTop + 100);

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
