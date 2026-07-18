<script lang="ts">
	import Image from "./Image.svelte";
	import MediaLightbox, {
		type LightboxItem,
	} from "./MediaLightbox.svelte";
	import type { Attachment } from "$lib/types/attachment";

	let {
		attachments,
		colsCount = 1,
		rowsCount = 1,
		imageFlex = 1,
	}: {
		attachments: Attachment[];
		colsCount?: number;
		rowsCount?: number;
		imageFlex?: number;
	} = $props();

	const items: LightboxItem[] = $derived(
		attachments.map((file) => ({
			link: file.link ?? "",
			name: file.name ?? "",
			type: file.type ?? "",
		})),
	);

	let open = $state(false);
	let index = $state(0);

	function openAt(i: number) {
		index = i;
		open = true;
	}
</script>

<div
	class={`grid grid-cols-${colsCount} grid-rows-${rowsCount} items-center gap-2 flex-${imageFlex} p-2 border-r-7`}
>
	{#each items as item, i (item.link || i)}
		<Image
			link={item.link}
			name={item.name}
			type={item.type ?? ""}
			onOpen={() => openAt(i)}
		/>
	{/each}
</div>

<MediaLightbox bind:open bind:index {items} />
