<script lang="ts">
	import MediaLightbox from "$lib/components/custom/MediaLightbox.svelte";
	import { isVideoAttachment } from "$lib/limits";

	let {
		link,
		name,
		type = "",
		onOpen,
	}: {
		link: string;
		name: string;
		type?: string;
		/** When set, parent owns the lightbox (e.g. AttachmentGallery). */
		onOpen?: () => void;
	} = $props();

	let open = $state(false);
	const isVideo = $derived(isVideoAttachment(name, type));
	const ownsLightbox = $derived(!onOpen);
</script>

<button
	type="button"
	class="flex cursor-pointer flex-col items-center justify-center gap-1 border-0 bg-transparent p-0 text-left"
	onclick={() => {
		if (onOpen) {
			onOpen();
		} else {
			open = true;
		}
	}}
>
	<span class="text-sm text-primary underline-offset-4 hover:underline">{name}</span>
	{#if isVideo}
		<!-- svelte-ignore a11y_media_has_caption -->
		<video
			src={link}
			muted
			playsinline
			preload="metadata"
			class="h-48 w-72 bg-black object-contain"
		></video>
	{:else}
		<img src={link} alt={name} class="h-48 w-72 object-scale-down" />
	{/if}
</button>

{#if ownsLightbox}
	<MediaLightbox
		bind:open
		items={[{ link, name, type }]}
		index={0}
	/>
{/if}
