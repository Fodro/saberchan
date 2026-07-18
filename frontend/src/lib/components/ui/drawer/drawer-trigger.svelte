<script lang="ts">
	import { Drawer as DrawerPrimitive } from "vaul-svelte";

	let { el = $bindable(undefined), ...restProps }: DrawerPrimitive.TriggerProps = $props();
</script>

<DrawerPrimitive.Trigger bind:el data-slot="drawer-trigger" {...restProps} />
