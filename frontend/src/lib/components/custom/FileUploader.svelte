<script lang="ts">
	import type { File as FileType } from "$lib/types/attachment";
	import { FilePlus, Trash } from "svelte-radix";
	import Button from "../ui/button/button.svelte";
	import { t } from "$lib/translations";
	import { toast } from "svelte-sonner";
	import { MAX_FILE_BYTES, MAX_FILES } from "$lib/limits";

	let { value = $bindable() }: { value: FileType[] } = $props();
	let fileCurrentId = $state(0);

	function onFilePicked(e: Event) {
		const input = e.target as HTMLInputElement | null;
		const picked = input?.files?.[0];
		if (!picked) return;

		const id = fileCurrentId.toString();
		fileCurrentId += 1;

		if (picked.size > MAX_FILE_BYTES) {
			toast.error($t("common.file.limitSize"));
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
	{#each value as file}
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
			input.accept = "image/*";
			input.onchange = onFilePicked;
			input.click();
		}}
	>
		<FilePlus />
	</Button>
</div>
