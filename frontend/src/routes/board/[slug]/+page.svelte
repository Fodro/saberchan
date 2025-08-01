<script lang="ts">
	import { invalidate } from "$app/navigation";
	import Draggable from "$lib/components/custom/Draggable.svelte";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import Button from "$lib/components/ui/button/button.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import Input from "$lib/components/ui/input/input.svelte";
	import Label from "$lib/components/ui/label/label.svelte";
	import { Separator } from "$lib/components/ui/separator/index.js";
	import { Textarea } from "$lib/components/ui/textarea/index.js";
	import Image from "$lib/components/custom/Image.svelte";
	import type { Attachment, File as FileType } from "$lib/types/attachment";
	import {
		bufferToBase64,
		formatDateTime,
		insertTagAtCursor,
	} from "$lib/helpers.js";
	import { trackBoard } from "$lib/tracking";
	import type { Thread } from "$lib/types/thread.js";
	import { getContext, onMount } from "svelte";
	import { t } from "$lib/translations";
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
	import PostBody from "$lib/components/custom/PostBody.svelte";
	import Captcha from "$lib/components/custom/Captcha.svelte";
	import { toast } from "svelte-sonner";
	import FileUploader from "$lib/components/custom/FileUploader.svelte";

	let { data } = $props();

	let counter: () => number = getContext("counter");

	let isReplyOpen = $state(false);

	let newTitle: string | null = $state(null);
	let newText: string | null = $state(null);

	let filesList: FileType[] = $state([]);

	let captchaInput = $state("");
	let captchaToken = $state("");

	const setCaptchaInput = (input: string) => {
		captchaInput = input;
	};

	const setCaptchaToken = (token: string) => {
		captchaToken = token;
	};

	$effect(() => {
		counter();
		invalidate("board:slug");
	});

	onMount(async () => {
		await trackBoard(data.board.alias);
	});
</script>

<svelte:head>
	<title>/{data.slug}/ - {data.board.name}</title>
</svelte:head>

<div class="flex flex-col justify-center items-start gap-5">
	<h3 class="mt-8 scroll-m-20 text-2xl font-semibold tracking-tight mb-5">
		{data.board.name}
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
			{$t("common.threads.new")}
		{/if}
	</Button>
</div>
<Separator class="my-4" />

{#if isReplyOpen}
	<Draggable>
		<Card.Root class="w-[100%] h-[95%]">
			<Card.Header>
				<Card.Title>{$t("common.threads.new")}</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="grid grid-cols-1 w-full items-center gap-2">
					<div class="flex flex-col justify-start items-start gap-3">
						<Label>{$t("common.fields.title")}</Label>
						<Input
							placeholder={$t("common.fields.title_placeholder")}
							bind:value={newTitle}
						/>
					</div>
					<div class="flex flex-col justify-start items-start gap-2">
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
										"new-thread-area",
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
										"new-thread-area",
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
										"new-thread-area",
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
										"new-thread-area",
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
										"new-thread-area",
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
										"new-thread-area",
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
										"new-thread-area",
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
								class="cursor-pointer"
								size="icon"
								variant="outline"
								on:click={() => {
									const field = document.getElementById(
										"new-thread-area",
									) as HTMLTextAreaElement;
									insertTagAtCursor(field, "\n>", "\n");
								}}
							>
								<CaretRight />
							</Button>
						</div>
						<Textarea
							id="new-thread-area"
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

							const res = await fetch("/api/thread", {
								method: "POST",
								body: JSON.stringify({
									board_id: data.board.id,
									title: newTitle,
									original_post: {
										text: newText,
										attachments: attachments,
									},
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
							const thread: Thread = await res.json();
							newText = null;
							newTitle = null;
							isReplyOpen = false;
							filesList = [];

							await window.open(
								`/board/${data.slug}/thread/${thread.id}`,
								"_blank",
								"noopener",
							);
						}}
					>
						{$t("common.post")}
					</Button>
				</div>
			</Card.Footer>
		</Card.Root>
	</Draggable>
{/if}

<div class="grid grid-cols-2 gap-4 pb-2">
	{#each data.board.threads as thread}
		<Card.Root>
			<Card.Header>
				<Card.Title>
					<div class="flex flex-row justify-start items-center gap-3">
						{thread.title}
						{#if thread.is_author}
							<Badge>{$t("common.you")}</Badge>
						{/if}
					</div>
				</Card.Title>
				<Card.Description>
					anon #{thread.original_post.number}, {$t(
						"common.posts.replies",
					)}: {thread.replies_count}
					{$t("common.posts.at")}
					{formatDateTime(thread.original_post.created_at)}
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="flex flex-row justify-start items-start gap-3">
					{#if thread.original_post.attachments && thread.original_post.attachments.length > 0}
						<div
							class={`grid grid-cols-1 grid-rows-1 items-center gap-2 flex-1 p-2 border-r-7 flex-1`}
						>
							<Image
								link={thread.original_post.attachments[0]
									.link ?? ""}
								name={thread.original_post.attachments[0]
									.name ?? ""}
							/>
						</div>
					{/if}
					<div class="flex-2">
						{#if thread.original_post.text.length <= 300}
							<PostBody text={thread.original_post.text} />
						{/if}
						{#if thread.original_post.text.length > 300}
							<p class="leading-7 whitespace-pre-wrap">
								{thread.original_post.text.substring(0, 300)}...
							</p>
						{/if}
					</div>
				</div>
			</Card.Content>
			<Card.Footer>
				<Button
					class="cursor-pointer"
					href={`/board/${data.slug}/thread/${thread.id}`}
					target="_blank"
					rel="noreferrer noopener"
				>
					{$t("common.reply")}
				</Button>
			</Card.Footer>
		</Card.Root>
	{/each}
</div>
