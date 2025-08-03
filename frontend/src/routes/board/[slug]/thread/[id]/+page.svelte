<script lang="ts">
	import Separator from "$lib/components/ui/separator/separator.svelte";
	import type { Attachment, File as FileType } from "$lib/types/attachment";
	import * as Card from "$lib/components/ui/card/index.js";
	import { bufferToBase64, insertTagAtCursor } from "$lib/helpers.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import { t } from "$lib/translations";
	import Draggable from "$lib/components/custom/Draggable.svelte";
	import { Textarea } from "$lib/components/ui/textarea/index.js";
	import { Label } from "$lib/components/ui/label/index.js";
	import { Checkbox } from "$lib/components/ui/checkbox/index.js";
	import { getContext } from "svelte";
	import { toast } from "svelte-sonner";
	import {
		FontBold,
		FontItalic,
		Underline,
		Overline,
		TextNone,
		CaretUp,
		CaretDown,
		TransparencyGrid,
		CaretRight,
		DrawingPinFilled,
		DrawingPin,
	} from "svelte-radix";
	import { invalidate } from "$app/navigation";
	import PostCard from "$lib/components/custom/PostCard.svelte";
	import { redirect } from "@sveltejs/kit";
	import Captcha from "$lib/components/custom/Captcha.svelte";
	import FileUploader from "$lib/components/custom/FileUploader.svelte";

	let newText = $state("");
	let newSage = $state(false);
	let newOP = $state(false);
	let isReplyOpen = $state(false);
	let captchaInput = $state("");
	let captchaToken = $state("");
	let filesList: FileType[] = $state([]);
	let counter: () => number = getContext("counter");

	let formX = $state(0);
	let formY = $state(0);
	let formPinned = $state(false);

	const { data } = $props();

	if (!data.thread) {
		redirect(302, "/404");
	}

	const checkIsInText = (txt: string): boolean => {
		return newText.includes(txt);
	};

	const addToText = (txt: string) => {
		newText += txt;
	};

	const setReplyOpen = (value: boolean) => {
		isReplyOpen = value;
	};

	const setCaptchaInput = (input: string) => {
		captchaInput = input;
	};

	const setCaptchaToken = (token: string) => {
		captchaToken = token;
	};

	$effect(() => {
		counter();
		invalidate("thread:id");
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

<svelte:window bind:scrollX={formX} bind:scrollY={formY} />

<div class="flex flex-col justify-center items-start gap-2">
	<h3 class="mt-8 scroll-m-20 text-2xl font-semibold tracking-tight mb-5">
		{data.thread.title}
	</h3>
	<Button
		class="cursor-pointer"
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
	<div class="grid grid-cols-1 gap-4 pb-2 w-[100%]">
		<PostCard
			post={data.thread.original_post}
			{addToText}
			{checkIsInText}
			{setReplyOpen}
			isSigned={data.signed}
		/>
		{#each data.thread.posts as post}
			<PostCard
				{post}
				{addToText}
				{checkIsInText}
				{setReplyOpen}
				isSigned={data.signed}
			/>
		{/each}
	</div>
</div>

{#if isReplyOpen}
	<Draggable initialLeft={formX} initialTop={formY} pinned={formPinned}>
		<Card.Root class="w-[100%] h-[100%]">
			<Card.Title>
				<div class="flex flex-row items-center h-[5%] flex-1 pl-6 pt-4 pr-6">
					<div class="flex flex-row flex-1">
						<p class="text-muted-foreground">
							{$t("common.posts.new")}
						</p>
					</div>
					<div class="flex flex-row-reverse flex-1">
						<Button
							class="cursor-pointer"
							variant="outline"
							size="icon"
							on:click={() => {
								formPinned = !formPinned;
							}}
						>
							{#if formPinned}
								<DrawingPinFilled />
							{/if}
							{#if !formPinned}
								<DrawingPin />
							{/if}
						</Button>
					</div>
				</div>
			</Card.Title>
			<Card.Description>
				<div class="flex flex-row items-center h-[5%] flex-1 pl-6">
					<p class="text-muted-foreground">
						{$t("common.draggable")}
					</p>
				</div>
			</Card.Description>
			<Card.Content>
				<div class="grid grid-cols-1 w-full items-center gap-4">
					<div class="flex flex-row justify-start items-center gap-2">
						<div
							class="flex flex-row justify-start items-center gap-2"
						>
							<Checkbox
								id="sage"
								bind:checked={newSage}
								class="cursor-pointer"
							/>
							<Label>
								{$t("common.fields.sage")}
							</Label>
						</div>
						{#if data.thread.original_post.is_author}
							<div
								class="flex flex-row justify-start items-center gap-2"
							>
								<Checkbox
									id="op"
									bind:checked={newOP}
									class="cursor-pointer"
								/>
								<Label>
									{$t("common.op")}
								</Label>
							</div>
						{/if}
					</div>
					<div class="flex flex-col justify-start items-start gap-3">
						<div
							class="flex flex-row justify-start items-center gap-2"
						>
							<Label>{$t("common.fields.text")}</Label>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[b]", "[/b]");
								}}
							>
								<FontBold />
							</Button>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[i]", "[/i]");
								}}
							>
								<FontItalic />
							</Button>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[u]", "[/u]");
								}}
							>
								<Underline />
							</Button>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[o]", "[/o]");
								}}
							>
								<Overline />
							</Button>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[s]", "[/s]");
								}}
							>
								<TextNone />
							</Button>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[sup]", "[/sup]");
								}}
							>
								<CaretUp />
							</Button>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "[sub]", "[/sub]");
								}}
							>
								<CaretDown />
							</Button>
							<Button
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
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
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-post-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "\n>", "\n");
								}}
							>
								<CaretRight />
							</Button>
						</div>
						<Textarea
							id="new-post-area"
							placeholder={$t("common.fields.text_placeholder")}
							rows={10}
							class="min-h-[70%] w-full resize-none"
							bind:value={newText}
						/>
						<FileUploader bind:value={filesList} />
						<Captcha {setCaptchaInput} {setCaptchaToken} />
					</div>
				</div>
			</Card.Content>
			<Card.Footer>
				<div
					class="flex flex-row justify-start items-center gap-4 w-full h-full"
				>
					<Button
						class="cursor-pointer"
						variant="secondary"
						on:click={() => {
							isReplyOpen = !isReplyOpen;
						}}
					>
						{$t("common.cancel")}
					</Button>
					<Button
						class="cursor-pointer"
						on:click={async () => {
							const attachments: Attachment[] = [];
							filesList.forEach((value) => {
								attachments.push({
									id: undefined,
									post_id: undefined,
									link: undefined,
									name: value.name,
									type: "image",
									body: bufferToBase64(value.blob),
								});
							});
							const res = await fetch("/api/post", {
								method: "POST",
								body: JSON.stringify({
									thread_id: data.thread.id,
									text: newText,
									sage: newSage,
									op_marker: newOP,
									attachments: attachments,
									captcha: {
										input: captchaInput,
										token: captchaToken,
									},
								}),
							});
							if (res.status != 201 && res.status != 200) {
								toast.error(await res.text());
								return;
							}

							newText = "";
							newSage = false;
							newOP = false;
							isReplyOpen = false;
							filesList = [];
						}}
					>
						{$t("common.post")}
					</Button>
				</div>
			</Card.Footer>
		</Card.Root>
	</Draggable>
{/if}
