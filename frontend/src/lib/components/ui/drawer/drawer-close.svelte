<script lang="ts">
	import { Drawer as DrawerPrimitive } from "vaul-svelte";

	let { el = $bindable(undefined), ...restProps }: DrawerPrimitive.CloseProps = $props();
</script>

<DrawerPrimitive.Close bind:el data-slot="drawer-close" {...restProps} />
