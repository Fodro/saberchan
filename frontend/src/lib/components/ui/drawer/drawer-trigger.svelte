<script lang="ts">
	import { Drawer as DrawerPrimitive } from "vaul-svelte";

	let { children, ...restProps }: DrawerPrimitive.TriggerProps = $props();
</script>

<DrawerPrimitive.Trigger data-slot="drawer-trigger" {...restProps}>
	{@render children?.()}
</DrawerPrimitive.Trigger>
