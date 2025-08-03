<script lang="ts">
    import type { Snippet } from "svelte";

	const {
		initialLeft,
		initialTop,
		pinned,
		children
	}: {
		initialLeft: number;
		initialTop: number;
		pinned: boolean;
		children: Snippet<[]>;
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
<section
	onmousedown={onMouseDown}
	style="left: {left}px; top: {top}px;"
	class="draggable w-[50vw] h-[70vh]"
>
	{@render children()}
</section>

<svelte:window on:mouseup={onMouseUp} on:mousemove={onMouseMove} />

<style>
	.draggable {
		user-select: none;
		cursor: move;
		position: absolute;
	}
</style>
