<script lang="ts">
	import type { File as FileType } from "$lib/types/attachment";
	import { FilePlus, Trash } from "svelte-radix";
	import Button from "../ui/button/button.svelte";
	import { t } from "$lib/translations";
	import { toast } from "svelte-sonner";
	import { nanoid } from "nanoid";

	let { value = $bindable() }: { value: FileType[] } = $props();
</script>

<div class="flex flex-row justify-start items-start gap-2">
	{#each value as file}
		<Button
			id={`file-${file.id}`}
			title={$t("common.file.remove")}
			on:click={async () => {
				value = value.filter((value) => value.id != file.id);
			}}
		>
			{file.name}
			<Trash />
		</Button>
	{/each}
	<Button
		size="icon"
		title={$t("common.file.add")}
		on:click={async () => {
			if (value.length >= 4) {
				toast.error($t("common.file.limitCount"));
				return;
			}

			const input = document.createElement("input");
			input.type = "file";
			input.accept = "image/*";

			let file: File;
			input.onchange = async (e: any) => {
				if (!e.target || !e.target.files) {
					return;
				}
				file = e.target.files[0];
				const id = nanoid(5);

				if (file.size > 2097152) {
					toast.error($t("common.file.limitSize"));
					return;
				}
				let name: string;
				if (file.name.length > 10) {
					const fileExt = file.name.split(".").pop();
					name = id + "." + fileExt;
				} else {
					name = file.name;
				}

				let reader = new FileReader();
				reader.readAsArrayBuffer(file);

				reader.onload = (readerEvent) => {
					if (!readerEvent.target) {
						toast.error("Something wrong with file");
						return
					}
					const content = readerEvent.target.result; // this is the content!
					value.push({
						id: id,
						name: name,
						blob: content,
					});
				};
			};
			input.click();
		}}
	>
		<FilePlus />
	</Button>
</div>
