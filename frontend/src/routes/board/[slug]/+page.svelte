<script lang="ts">
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import Button from "$lib/components/ui/button/button.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import Input from "$lib/components/ui/input/input.svelte";
	import Label from "$lib/components/ui/label/label.svelte";
	import { Separator } from "$lib/components/ui/separator/index.js";
	import ComposeForm from "$lib/components/custom/ComposeForm.svelte";
	import Image from "$lib/components/custom/Image.svelte";
	import type { File as FileType } from "$lib/types/attachment";
	import { formatDateTime } from "$lib/helpers.js";
	import { composeErrorMessageFactory, submitCompose } from "$lib/compose";
	import { trackBoard } from "$lib/tracking";
	import type { Thread } from "$lib/types/thread.js";
	import { onMount } from "svelte";
	import { t } from "$lib/translations";
	import PostBody from "$lib/components/custom/PostBody.svelte";
	import { toast } from "svelte-sonner";

	let { data } = $props();

	let isReplyOpen = $state(false);
	let submitting = $state(false);

	let formX = $state(0);
	let formY = $state(0);
	let formPinned = $state(true);

	let newTitle: string | null = $state(null);
	let newText = $state("");

	let filesList: FileType[] = $state([]);

	let captchaInput = $state("");
	let captchaToken = $state("");
	let captchaCounter = $state(0);

	const composeErrorMessage = composeErrorMessageFactory((key) => $t(key));

	onMount(async () => {
		await trackBoard(data.board.alias);
	});

	const submitNewThread = async () => {
		submitting = true;
		try {
			const result = await submitCompose({
				endpoint: "/api/thread",
				fields: {
					board_id: data.board.id,
					title: newTitle ?? "",
					text: newText,
				},
				title: newTitle,
				text: newText,
				requireTitle: true,
				captchaInput,
				captchaToken,
				files: filesList,
				errorMessage: composeErrorMessage,
				captchaFailedMessage: $t("common.captcha.failed"),
			});

			if (!result.ok) {
				toast.error(result.message);
				if (result.bumpCaptcha) captchaCounter += 1;
				return;
			}

			const thread = result.json as Thread;
			newText = "";
			newTitle = null;
			isReplyOpen = false;
			filesList = [];

			await window.open(`/board/${data.slug}/thread/${thread.id}`, "_blank", "noopener");
		} finally {
			submitting = false;
		}
	};
</script>

<svelte:head>
	<title>/{data.slug}/ - {data.board.name}</title>
</svelte:head>

<svelte:window bind:scrollX={formX} bind:scrollY={formY} />

<div class="flex flex-col justify-center items-start gap-5">
	<h3 class="mt-8 scroll-m-20 text-2xl font-semibold tracking-tight mb-5">
		{data.board.name}
	</h3>
	<Button
		class="cursor-pointer"
		onclick={() => {
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
	<ComposeForm
		heading={$t("common.threads.new")}
		textareaId="new-thread-area"
		initialLeft={formX}
		initialTop={formY}
		bind:formPinned
		bind:text={newText}
		bind:files={filesList}
		bind:captchaInput
		bind:captchaToken
		{captchaCounter}
		onCancel={() => {
			isReplyOpen = false;
		}}
		onSubmit={submitNewThread}
		{submitting}
	>
		{#snippet beforeText()}
			<div class="flex flex-col justify-start items-start gap-3">
				<Label>{$t("common.fields.title")}</Label>
				<Input placeholder={$t("common.fields.title_placeholder")} bind:value={newTitle} />
			</div>
		{/snippet}
	</ComposeForm>
{/if}

<div class="grid md:grid-cols-2 grid-cols-1 gap-4 pb-2">
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
							class="grid grid-cols-1 grid-rows-1 items-center gap-2 flex-1 p-2 border-r-7 flex-1"
						>
							<Image
								link={thread.original_post.attachments[0].link ?? ""}
								name={thread.original_post.attachments[0].name ?? ""}
							/>
						</div>
					{/if}
					<div class="flex-2">
						{#if thread.original_post.text.length <= 300}
							<PostBody text={thread.original_post.text} additionalClass="" />
						{/if}
						{#if thread.original_post.text.length > 300}
							<PostBody
								text={thread.original_post.text.substring(0, 300) + "..."}
								additionalClass="leading-7 whitespace-pre-wrap"
							/>
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

{#if (data.board.total_threads ?? 0) > (data.board.limit ?? 20)}
	{@const pageLimit = data.board.limit ?? 20}
	{@const pageOffset = data.board.offset ?? 0}
	{@const totalThreads = data.board.total_threads ?? 0}
	{@const page = Math.floor(pageOffset / pageLimit) + 1}
	{@const totalPages = Math.max(1, Math.ceil(totalThreads / pageLimit))}
	{@const prevOffset = Math.max(0, pageOffset - pageLimit)}
	{@const nextOffset = pageOffset + pageLimit}
	<div class="flex flex-row justify-center items-center gap-3 w-full py-4">
		{#if pageOffset > 0}
			<Button
				variant="outline"
				href={`/board/${data.slug}?limit=${pageLimit}&offset=${prevOffset}`}
			>
				←
			</Button>
		{/if}
		<span class="text-muted-foreground text-sm">{page} / {totalPages}</span>
		{#if nextOffset < totalThreads}
			<Button
				variant="outline"
				href={`/board/${data.slug}?limit=${pageLimit}&offset=${nextOffset}`}
			>
				→
			</Button>
		{/if}
	</div>
{/if}
