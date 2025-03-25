<script lang="ts">
	import Separator from "$lib/components/ui/separator/separator.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import { formatDateTime, insertTagAtCursor, scrollIntoView } from "$lib/helpers.js";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import DoubleArrowDown from "svelte-radix/DoubleArrowDown.svelte";
	import { t } from "$lib/translations";
	import Draggable from "$lib/components/custom/Draggable.svelte";
	import { Textarea } from "$lib/components/ui/textarea/index.js";
	import { Label } from "$lib/components/ui/label/index.js";
	import { Checkbox } from "$lib/components/ui/checkbox/index.js";
	import { onMount } from "svelte";
	import { toast } from "svelte-sonner";
    import { FontBold, FontItalic, Underline, Overline, TextNone, CaretUp, CaretDown, TransparencyGrid, CaretRight } from "svelte-radix";

	let newText = $state("");
	let newSage = $state(false);
	let isReplyOpen = $state(false);
	let hash = $state("");

	const { data } = $props();

	$effect(() => {
		scrollIntoView(hash);
	});

	onMount(() => {
		hash = window.location.hash;
	});
</script>

<svelte:head>
	<title>
		/{data.slug}/ - #{data.thread.original_post.number} - {data.thread.original_post.text.substring(
			0,
			15,
		)}...
	</title>
</svelte:head>

<div class="flex flex-col justify-center items-start gap-2">
	<h3 class="mt-8 scroll-m-20 text-2xl font-semibold tracking-tight mb-5">
		{data.thread.title}
	</h3>
	<Button
		on:click={() => {
			isReplyOpen = !isReplyOpen;
		}}
	>
		{#if isReplyOpen}
			{$t("common.cancel")}
		{/if}
		{#if !isReplyOpen}
			{$t("common.reply")}
		{/if}
	</Button>
	<Separator />
	<div class="grid grid-cols-1 gap-4">
		<Card.Root id={`#${data.thread.original_post.number}`} class={`${hash === `#${data.thread.original_post.number}` ? 'border-sky-500' : ''} ${data.thread.original_post.is_author ? 'outline' : ''}`}>
			<Card.Header>
				<Card.Title
					>anon #{data.thread.original_post.number}
					{$t("common.posts.at")}
					{formatDateTime(data.thread.original_post.created_at)}
				</Card.Title>
				<div class="flex flex-row justify-start items-center gap-2">
					{#if data.thread.original_post.is_author}
						<Badge>{$t("common.you")}</Badge>
					{/if}
					<Badge>{$t("common.op")}</Badge>
				</div>
			</Card.Header>
			<Card.Content>
				<div class="flex flex-col justify-center items-start gap-2">
					<p class="leading-7 whitespace-pre-wrap">
						{data.thread.original_post.text}
					</p>
				</div>
			</Card.Content>
			<Card.Footer>
				<div
					class="flex flex-row justify-start items-center gap-4 w-full h-full"
				>
					<Button
						variant="secondary"
						on:click={() => {
							const toAppend = `>>${data.thread.original_post.number}\n`;
							if (!newText.includes(toAppend)) {
								newText += toAppend;
							}
							isReplyOpen = true;
						}}
					>
						{$t("common.reply")}
					</Button>
					<Button
						variant="outline"
						on:click={async () => {
							const link = `${window.location.href}#${data.thread.original_post.number}`;
							await navigator.clipboard.writeText(link);
							toast.success($t("common.copied"));
						}}
					>
						{$t("common.copy_link")}
					</Button>
				</div>
			</Card.Footer>
		</Card.Root>
		{#each data.thread.posts as post}
			<Card.Root id={`#${post.number}`} class={`${hash === `#${post.number}` ? 'border-sky-500' : ''} ${post.is_author ? 'outline' : ''}`}>
				<Card.Header>
					<Card.Title
						>anon #{post.number}
						{$t("common.posts.at")}
						{formatDateTime(post.created_at)}
					</Card.Title>
					<div class="flex flex-row justify-start items-center gap-2">
						{#if post.is_author}
							<Badge>{$t("common.you")}</Badge>
						{/if}
						{#if post.op_marker}
							<Badge>{$t("common.op")}</Badge>
						{/if}
						{#if post.sage}
							<Badge variant="destructive">
								<DoubleArrowDown
									class="h-[1.2rem] w-[1.2rem]"
								/>
							</Badge>
						{/if}
					</div>
				</Card.Header>
				<Card.Content>
					<div class="flex flex-col justify-center items-start gap-2">
						<p class="leading-7 whitespace-pre-wrap">
							{post.text}
						</p>
					</div>
				</Card.Content>
				<Card.Footer>
					<div
						class="flex flex-row justify-start items-center gap-4 w-full h-full"
					>
						<Button
							variant="secondary"
							on:click={() => {
								const toAppend = `>>${post.number}\n`;
								if (!newText.includes(toAppend)) {
									newText += toAppend;
								}
								isReplyOpen = true;
							}}
						>
							{$t("common.reply")}
						</Button>
						<Button
							variant="outline"
							on:click={async () => {
								const base = window.location.href.split('#');
								const link = `${base[0]}#${post.number}`;
								await navigator.clipboard.writeText(link);
								toast.success($t("common.copied"));
							}}
						>
							{$t("common.copy_link")}
						</Button>
					</div>
				</Card.Footer>
			</Card.Root>
		{/each}
	</div>
</div>

{#if isReplyOpen}
	<Draggable>
		<Card.Root class="w-[50vw] h-[50vh]">
			<Card.Header>
				<Card.Title>{$t("common.posts.new")}</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="grid grid-cols-1 w-full items-center gap-4">
					<div class="flex flex-row justify-start items-center gap-2">
						<Checkbox id="sage" bind:checked={newSage} />
						<Label>
							{$t("common.fields.sage")}
						</Label>
					</div>
					<div class="flex flex-col justify-start items-start gap-3">
						<div
							class="flex flex-row justify-start items-center gap-2"
						>
							<Label>{$t("common.fields.text")}</Label>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[b]", "[/b]");
								}}
							>
								<FontBold />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[i]", "[/i]");
								}}
							>
								<FontItalic />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[u]", "[/u]");
								}}
							>
								<Underline />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[o]", "[/o]");
								}}
							>
								<Overline />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[s]", "[/s]");
								}}
							>
								<TextNone />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[sup]", "[/sup]");
								}}
							>
								<CaretUp />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[sub]", "[/sub]");
								}}
							>
								<CaretDown />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(
										field,
										"[spoiler]",
										"[/spoiler]",
									);
								}}
							>
								<TransparencyGrid />
							</Button>
							<Button
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(
										field,
										"\n>",
										"\n",
									);
								}}
							>
								<CaretRight />
							</Button>
						</div>
						<Textarea
							placeholder={$t("common.fields.text_placeholder")}
							rows={10}
							class="min-h-[70%] w-full resize-none"
							bind:value={newText}
						/>
					</div>
				</div>
			</Card.Content>
			<!-- TODO: add captcha -->
			<Card.Footer>
				<div
					class="flex flex-row justify-start items-center gap-4 w-full h-full"
				>
					<Button
						variant="secondary"
						on:click={() => {
							isReplyOpen = !isReplyOpen;
						}}
					>
						{$t("common.cancel")}
					</Button>
					<Button
						on:click={async () => {
							newText = "";
							isReplyOpen = false;
						}}
					>
						{$t("common.post")}
					</Button>
				</div>
			</Card.Footer>
		</Card.Root>
	</Draggable>
{/if}
