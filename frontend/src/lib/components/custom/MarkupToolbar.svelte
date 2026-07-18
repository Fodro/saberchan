<script lang="ts">
	import Button from "$lib/components/ui/button/button.svelte";
	import { insertTagAtCursor } from "$lib/helpers.js";
	import {
		CaretDown,
		CaretRight,
		CaretUp,
		FontBold,
		FontItalic,
		Overline,
		TextNone,
		TransparencyGrid,
		Underline,
	} from "svelte-radix";

	let { textareaId }: { textareaId: string } = $props();

	function apply(open: string, close: string) {
		const field = document.getElementById(textareaId) as HTMLTextAreaElement | null;
		if (!field) return;
		insertTagAtCursor(field, open, close);
	}
</script>

<div class="md:flex md:flex-row md:justify-start md:items-center grid grid-cols-3 gap-2">
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("[b]", "[/b]")}>
		<FontBold />
	</Button>
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("[i]", "[/i]")}>
		<FontItalic />
	</Button>
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("[u]", "[/u]")}>
		<Underline />
	</Button>
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("[o]", "[/o]")}>
		<Overline />
	</Button>
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("[s]", "[/s]")}>
		<TextNone />
	</Button>
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("[sup]", "[/sup]")}>
		<CaretUp />
	</Button>
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("[sub]", "[/sub]")}>
		<CaretDown />
	</Button>
	<Button
		class="cursor-pointer"
		size="icon"
		variant="outline"
		onclick={() => apply("[spoiler]", "[/spoiler]")}
	>
		<TransparencyGrid />
	</Button>
	<Button class="cursor-pointer" size="icon" variant="outline" onclick={() => apply("\n>", "\n")}>
		<CaretRight />
	</Button>
</div>
