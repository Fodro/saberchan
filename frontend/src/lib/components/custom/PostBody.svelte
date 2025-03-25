<script lang="ts">
	import { goto, invalidateAll } from "$app/navigation";
	import { Render } from "@jill64/svelte-sanitize";

	const sanitizeOpts = {
		sanitizeHtml: {
			allowedTags: ["span"],
			allowedAttributes: {
				span: ["class"],
			},
			allowedClasses: {
				span: [
					"text-xs",
					"align-sub",
					"bg-zinc-500",
					"hover:text-white",
					"text-zinc-500",
					"align-super",
					"line-through",
					"overline",
					"underline",
					"italic",
					"font-bold",
				],
			},
		},
	};

	let { text }: { text: string } = $props();
	let styledText = text
		.replaceAll("[b]", `<span class="font-bold">`)
		.replaceAll("[/b]", "</span>")
		.replaceAll("[i]", `<span class="italic">`)
		.replaceAll("[/i]", "</span>")
		.replaceAll("[u]", `<span class="underline">`)
		.replaceAll("[/u]", "</span>")
		.replaceAll("[o]", `<span class="overline">`)
		.replaceAll("[/o]", "</span>")
		.replaceAll("[s]", `<span class="line-through">`)
		.replaceAll("[/s]", "</span>")
		.replaceAll("[sup]", `<span class="text-xs align-super">`)
		.replaceAll("[/sup]", "</span>")
		.replaceAll(
			"[spoiler]",
			`<span class="bg-zinc-500 text-zinc-500 hover:text-white">`,
		)
		.replaceAll("[/spoiler]", "</span>")
		.replaceAll("[sub]", `<span class="text-xs align-sub">`)
		.replaceAll("[/sub]", "</span>");

	let lines = styledText.split("\n");
</script>

{#each lines as line}
	{#if line.trim().charAt(0) === ">" && line.trim().charAt(1) === ">"}
		<!-- svelte-ignore event_directive_deprecated -->
		<a
			class="text-orange-500 hover:underline cursor-pointer"
			href={`#${line.trim().replaceAll(">>", "")}`}
			data-sveltekit-reload
		>
			<Render html={line.trim()} options={sanitizeOpts} />
		</a>
	{:else if line.trim().charAt(0) === ">"}
		<p class="text-green-500">
			<Render html={line.trim()} options={sanitizeOpts} />
		</p>
	{:else}
		<p>
			<Render html={line.trim()} options={sanitizeOpts} />
		</p>
	{/if}
{/each}
