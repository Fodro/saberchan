<script lang="ts">
	import { untrack, type Snippet } from "svelte";

	const {
		initialLeft,
		initialTop,
		pinned,
		children,
	}: {
		initialLeft: number;
		initialTop: number;
		pinned: boolean;
		children: Snippet<[]>;
	} = $props();

	let leftMod = 300;
	let topMod = 100;

	if (typeof window !== "undefined") {
		leftMod = window.innerWidth / 4;
		topMod = window.innerHeight / 4;
		if (window.innerWidth < 770) {
			leftMod = 0;
			topMod = 0;
		}
	}

	// Seed once from props + viewport offset; drag mutates thereafter (not reactive to prop changes)
	let left = $state(untrack(() => initialLeft + leftMod));
	let top = $state(untrack(() => initialTop + topMod));

	let moving = $state(false);

	function onMouseDown() {
		if (pinned) {
			return;
		}
		moving = true;
	}

	function onMouseMove(e: MouseEvent) {
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
	class={`md:draggable absolute w-[100%] md:w-[50vw] md:h-[70vh]${pinned ? "" : " cursor-move"}`}
>
	{@render children()}
</section>

<svelte:window onmouseup={onMouseUp} onmousemove={onMouseMove} />

<style>
	.draggable {
		user-select: none;
	}
</style>
