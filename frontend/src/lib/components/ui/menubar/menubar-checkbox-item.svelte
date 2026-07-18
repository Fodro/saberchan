<script lang="ts">
	import { Menubar as MenubarPrimitive } from "bits-ui";
	import Check from "svelte-radix/Check.svelte";
	import Minus from "svelte-radix/Minus.svelte";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		checked = $bindable(false),
		indeterminate = $bindable(false),
		children: childrenProp,
		...restProps
	}: MenubarPrimitive.CheckboxItemProps = $props();
</script>

<MenubarPrimitive.CheckboxItem
	bind:ref
	bind:checked
	bind:indeterminate
	class={cn(
		"data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground relative flex cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
		className
	)}
	{...restProps}
>
	{#snippet children({ checked, indeterminate })}
		<span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			{#if indeterminate}
				<Minus class="h-4 w-4" />
			{:else}
				<Check class={cn("h-4 w-4", !checked && "text-transparent")} />
			{/if}
		</span>
		{@render childrenProp?.({ checked, indeterminate })}
	{/snippet}
</MenubarPrimitive.CheckboxItem>
