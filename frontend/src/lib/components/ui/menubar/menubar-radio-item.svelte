<script lang="ts">
	import { Menubar as MenubarPrimitive } from "bits-ui";
	import DotFilled from "svelte-radix/DotFilled.svelte";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		value,
		children: childrenProp,
		...restProps
	}: MenubarPrimitive.RadioItemProps = $props();
</script>

<MenubarPrimitive.RadioItem
	bind:ref
	{value}
	class={cn(
		"data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground relative flex cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
		className
	)}
	{...restProps}
>
	{#snippet children({ checked })}
		<span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			{#if checked}
				<DotFilled class="h-4 w-4 fill-current" />
			{/if}
		</span>
		{@render childrenProp?.({ checked })}
	{/snippet}
</MenubarPrimitive.RadioItem>
