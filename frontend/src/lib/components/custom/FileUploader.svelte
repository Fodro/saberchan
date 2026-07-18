<script lang="ts">
	import type { File as FileType } from "$lib/types/attachment";
	import { FilePlus, Trash } from "svelte-radix";
	import Button from "../ui/button/button.svelte";
	import { t } from "$lib/translations";
	import { toast } from "svelte-sonner";
	import {
		IMAGE_MIME,
		MAX_FILES,
		VIDEO_MIME,
		maxBytesForMime,
	} from "$lib/limits";

	let { value = $bindable() }: { value: FileType[] } = $props();
	let fileCurrentId = $state(0);

	function onFilePicked(e: Event) {
		const input = e.target as HTMLInputElement | null;
		const picked = input?.files?.[0];
		if (!picked) return;

		const id = fileCurrentId.toString();
		fileCurrentId += 1;

		const mime = picked.type || "";
		const allowed = IMAGE_MIME.has(mime) || VIDEO_MIME.has(mime);
		if (mime && !allowed) {
			toast.error($t("common.file.limitType"));
			return;
		}

		const limit = maxBytesForMime(mime || "image/jpeg");
		if (picked.size > limit) {
			toast.error(
				VIDEO_MIME.has(mime)
					? $t("common.file.limitSizeVideo")
					: $t("common.file.limitSize"),
			);
			return;
		}

		let fileExt = picked.name.split(".").pop() || "jpg";
		const name =
			picked.name.length > 10 + fileExt.length
				? picked.name.substring(0, 8) + "." + fileExt
				: picked.name;

		const reader = new FileReader();
		reader.readAsArrayBuffer(picked);
		reader.onload = (readerEvent) => {
			if (!readerEvent.target?.result) {
				toast.error("Something wrong with file");
				return;
			}
			value.push({
				id,
				name,
				blob: readerEvent.target.result,
			});
		};
	}
</script>

<div class="flex flex-row justify-start items-start gap-2">
	{#each value as file (file.id)}
		<Button
			id={`file-${file.id}`}
			title={$t("common.file.remove")}
			onclick={() => {
				value = value.filter((v) => v.id != file.id);
			}}
		>
			{file.name}
			<Trash />
		</Button>
	{/each}
	<Button
		size="icon"
		title={$t("common.file.add")}
		onclick={() => {
			if (value.length >= MAX_FILES) {
				toast.error($t("common.file.limitCount"));
				return;
			}

			const input = document.createElement("input");
			input.type = "file";
			input.accept = "image/*,video/webm,video/mp4";
			input.onchange = onFilePicked;
			input.click();
		}}
	>
		<FilePlus />
	</Button>
</div>
