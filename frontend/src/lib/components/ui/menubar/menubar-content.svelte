<script lang="ts">
	import { Menubar as MenubarPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		align = "start",
		alignOffset = -4,
		sideOffset = 8,
		children,
		...restProps
	}: MenubarPrimitive.ContentProps = $props();
</script>

<MenubarPrimitive.Portal>
	<MenubarPrimitive.Content
		bind:ref
		{align}
		{alignOffset}
		{sideOffset}
		class={cn(
			"bg-popover text-popover-foreground z-50 min-w-[12rem] rounded-md border p-1 shadow-md focus:outline-none",
			className
		)}
		{...restProps}
	>
		{@render children?.()}
	</MenubarPrimitive.Content>
</MenubarPrimitive.Portal>
