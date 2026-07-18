<script lang="ts">
	import { Dialog as DialogPrimitive } from "bits-ui";
	import Cross2 from "svelte-radix/Cross2.svelte";
	import { t } from "$lib/translations";

	let {
		open = $bindable(false),
		link,
		name,
	}: {
		open?: boolean;
		link: string;
		name: string;
	} = $props();
</script>

<DialogPrimitive.Root bind:open>
	<DialogPrimitive.Portal>
		<DialogPrimitive.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/75 duration-150"
		/>
		<DialogPrimitive.Content
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed left-1/2 top-1/2 z-50 flex max-h-[92vh] max-w-[92vw] -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-3 border-0 bg-transparent p-0 shadow-none outline-none duration-150"
		>
			<div class="flex w-full max-w-[90vw] items-center justify-between gap-3">
				<DialogPrimitive.Title
					class="truncate text-sm font-medium text-white/90"
				>
					{name}
				</DialogPrimitive.Title>
				<DialogPrimitive.Close
					class="shrink-0 rounded-sm text-white/80 transition-opacity duration-150 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white/60"
					aria-label={$t("common.lightbox.close")}
				>
					<Cross2 class="h-6 w-6" />
				</DialogPrimitive.Close>
			</div>
			<DialogPrimitive.Description class="sr-only">
				{$t("common.lightbox.viewing")}
			</DialogPrimitive.Description>
			<img
				src={link}
				alt={name}
				class="max-h-[84vh] max-w-[90vw] object-contain"
			/>
		</DialogPrimitive.Content>
	</DialogPrimitive.Portal>
</DialogPrimitive.Root>
