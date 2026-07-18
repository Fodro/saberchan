<script lang="ts">
	import { Drawer as DrawerPrimitive } from "vaul-svelte";
	import { cn } from "$lib/utils.js";

	let {
		el = $bindable(undefined),
		class: className,
		...restProps
	}: DrawerPrimitive.DescriptionProps = $props();
</script>

<DrawerPrimitive.Description
	bind:el
	data-slot="drawer-description"
	class={cn("text-muted-foreground text-sm", className)}
	{...restProps}
/>
