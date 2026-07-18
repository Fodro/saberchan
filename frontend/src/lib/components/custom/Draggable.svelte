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

	// Pinned uses fixed (viewport) coords; unpinned uses absolute (document) coords.
	// Seed once from props + viewport offset; drag mutates thereafter (not reactive to prop changes).
	let left = $state(untrack(() => (pinned ? leftMod : initialLeft + leftMod)));
	let top = $state(untrack(() => (pinned ? topMod : initialTop + topMod)));

	let el = $state<HTMLElement | undefined>();
	let moving = $state(false);

	/** Call before flipping `pinned` so left/top match the upcoming position mode. */
	export function preparePinChange(nextPinned: boolean) {
		if (!el || nextPinned === pinned) {
			return;
		}
		const rect = el.getBoundingClientRect();
		if (nextPinned) {
			left = rect.left;
			top = rect.top;
		} else {
			left = rect.left + window.scrollX;
			top = rect.top + window.scrollY;
		}
	}

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
	bind:this={el}
	onmousedown={onMouseDown}
	style="left: {left}px; top: {top}px;"
	class={`md:draggable z-50 w-[100%] md:w-[50vw] md:h-[70vh] ${pinned ? "fixed" : "absolute cursor-move"}`}
>
	{@render children()}
</section>

<svelte:window onmouseup={onMouseUp} onmousemove={onMouseMove} />

<style>
	.draggable {
		user-select: none;
	}
</style>
