<script lang="ts">
	import { Drawer as DrawerPrimitive } from "vaul-svelte";
	import { cn } from "$lib/utils.js";

	let {
		el = $bindable(undefined),
		class: className,
		...restProps
	}: DrawerPrimitive.TitleProps = $props();
</script>

<DrawerPrimitive.Title
	bind:el
	data-slot="drawer-title"
	class={cn("text-foreground font-semibold", className)}
	{...restProps}
/>
