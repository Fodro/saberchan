<script lang="ts">
	import { Dialog as DialogPrimitive } from "bits-ui";
	import Cross2 from "svelte-radix/Cross2.svelte";
	import ChevronLeft from "svelte-radix/ChevronLeft.svelte";
	import ChevronRight from "svelte-radix/ChevronRight.svelte";
	import { t } from "$lib/translations";
	import { isVideoAttachment } from "$lib/limits";

	export type LightboxItem = {
		link: string;
		name: string;
		type?: string;
	};

	let {
		open = $bindable(false),
		index = $bindable(0),
		items,
	}: {
		open?: boolean;
		index?: number;
		items: LightboxItem[];
	} = $props();

	const hasGallery = $derived(items.length > 1);
	const safeIndex = $derived(
		items.length === 0 ? 0 : ((index % items.length) + items.length) % items.length,
	);
	const current = $derived(items[safeIndex]);
	const isVideo = $derived(
		current ? isVideoAttachment(current.name, current.type ?? "") : false,
	);
	let videoEl: HTMLVideoElement | undefined = $state();

	function pauseVideo() {
		videoEl?.pause();
	}

	function goPrev() {
		if (!hasGallery) return;
		pauseVideo();
		index = (safeIndex - 1 + items.length) % items.length;
	}

	function goNext() {
		if (!hasGallery) return;
		pauseVideo();
		index = (safeIndex + 1) % items.length;
	}

	function onKeydown(e: KeyboardEvent) {
		if (!open || !hasGallery) return;
		if (e.key === "ArrowLeft") {
			e.preventDefault();
			goPrev();
		} else if (e.key === "ArrowRight") {
			e.preventDefault();
			goNext();
		}
	}

	$effect(() => {
		if (!open) {
			pauseVideo();
		}
	});
</script>

<svelte:window onkeydown={onKeydown} />

<DialogPrimitive.Root bind:open>
	<DialogPrimitive.Portal>
		<DialogPrimitive.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/75 duration-150"
		/>
		<DialogPrimitive.Content
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed left-1/2 top-1/2 z-50 flex max-h-[92vh] max-w-[92vw] -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-3 border-0 bg-transparent p-0 shadow-none outline-none duration-150"
		>
			{#if current}
				<div class="flex w-full max-w-[90vw] items-center justify-between gap-3">
					<div class="flex min-w-0 items-center gap-2">
						<DialogPrimitive.Title
							class="truncate text-sm font-medium text-white/90"
						>
							{current.name}
						</DialogPrimitive.Title>
						{#if hasGallery}
							<span class="shrink-0 text-sm text-white/60">
								{safeIndex + 1} / {items.length}
							</span>
						{/if}
					</div>
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
				<div class="flex items-center gap-2">
					{#if hasGallery}
						<button
							type="button"
							class="shrink-0 rounded-sm p-2 text-white/80 transition-opacity duration-150 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white/60"
							aria-label={$t("common.lightbox.prev")}
							onclick={(e) => {
								e.stopPropagation();
								goPrev();
							}}
						>
							<ChevronLeft class="h-8 w-8" />
						</button>
					{/if}
					{#key safeIndex}
						{#if isVideo}
							<!-- svelte-ignore a11y_media_has_caption -->
							<video
								bind:this={videoEl}
								src={current.link}
								controls
								playsinline
								preload="metadata"
								class="max-h-[84vh] bg-black {hasGallery
									? 'max-w-[80vw]'
									: 'max-w-[90vw]'}"
							></video>
						{:else}
							<img
								src={current.link}
								alt={current.name}
								class="max-h-[84vh] object-contain {hasGallery
									? 'max-w-[80vw]'
									: 'max-w-[90vw]'}"
							/>
						{/if}
					{/key}
					{#if hasGallery}
						<button
							type="button"
							class="shrink-0 rounded-sm p-2 text-white/80 transition-opacity duration-150 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white/60"
							aria-label={$t("common.lightbox.next")}
							onclick={(e) => {
								e.stopPropagation();
								goNext();
							}}
						>
							<ChevronRight class="h-8 w-8" />
						</button>
					{/if}
				</div>
			{/if}
		</DialogPrimitive.Content>
	</DialogPrimitive.Portal>
</DialogPrimitive.Root>
