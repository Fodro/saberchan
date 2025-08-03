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

	let leftMod = 300;
	let topMod = 100;

	if (window) {
		leftMod = window.innerWidth / 4
		topMod = window.innerHeight / 4
	}

	let left = $state(initialLeft + leftMod);
	let top = $state(initialTop + topMod);

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
	class={`draggable w-[50vw] h-[70vh]${pinned? '' : ' cursor-move'}`}
>
	{@render children()}
</section>

<svelte:window on:mouseup={onMouseUp} on:mousemove={onMouseMove} />

<style>
	.draggable {
		user-select: none;
		position: absolute;
	}
</style>
