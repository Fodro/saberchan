<script lang="ts">
	import { Menubar as MenubarPrimitive } from "bits-ui";
	import type { HTMLAnchorAttributes } from "svelte/elements";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		inset,
		href = undefined,
		target = undefined,
		rel = undefined,
		children,
		...restProps
	}: MenubarPrimitive.ItemProps & {
		inset?: boolean;
		href?: string | undefined;
		target?: HTMLAnchorAttributes["target"];
		rel?: HTMLAnchorAttributes["rel"];
	} = $props();

	const itemClass =
		"data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50";
</script>

{#if href}
	<MenubarPrimitive.Item bind:ref {...restProps}>
		{#snippet child({ props })}
			<a {href} {target} {rel} class={cn(itemClass, inset && "pl-8", className)} {...props}>
				{@render children?.()}
			</a>
		{/snippet}
	</MenubarPrimitive.Item>
{:else}
	<MenubarPrimitive.Item bind:ref class={cn(itemClass, inset && "pl-8", className)} {...restProps}>
		{@render children?.()}
	</MenubarPrimitive.Item>
{/if}
