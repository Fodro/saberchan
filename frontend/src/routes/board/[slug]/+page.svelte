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
	import { followThread, unfollowThread } from "$lib/followed";
	import { trackBoard } from "$lib/tracking";
	import type { Thread } from "$lib/types/thread.js";
	import { onMount } from "svelte";
	import { t } from "$lib/translations";
	import PostBody from "$lib/components/custom/PostBody.svelte";
	import { toast } from "svelte-sonner";
	import { restoreDeleted, softDelete } from "$lib/adminModeration";
	import { Trash, Update } from "svelte-radix";
	import { invalidate } from "$app/navigation";

	let { data } = $props();

	let isReplyOpen = $state(false);
	let submitting = $state(false);
	let modBusy = $state(false);
	let followBusyId = $state<string | null>(null);

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
	const canCreateThread = $derived(!data.board.locked || data.isAdmin);
	const isBoardDeleted = $derived(Boolean(data.board.deleted_at));
	const followedIds = $derived(new Set(data.followed?.ids ?? []));

	onMount(async () => {
		await trackBoard(data.board.alias);
	});

	async function toggleFollowCatalog(thread: Thread) {
		followBusyId = thread.id;
		const wasFollowing = followedIds.has(thread.id);
		try {
			if (wasFollowing) {
				await unfollowThread(thread.id);
			} else {
				await followThread(thread.id, thread.replies_count ?? 0);
			}
		} catch {
			if (!wasFollowing) {
				toast.error($t("common.follow_failed"));
			}
		} finally {
			followBusyId = null;
		}
	}

	async function deleteBoard() {
		modBusy = true;
		try {
			const err = await softDelete("board", data.board.id);
			if (err) {
				toast.error(err);
				return;
			}
			await invalidate("board:slug");
			await invalidate("board:all");
		} finally {
			modBusy = false;
		}
	}

	async function restoreBoard() {
		modBusy = true;
		try {
			const err = await restoreDeleted("board", data.board.id);
			if (err) {
				toast.error(err);
				return;
			}
			await invalidate("board:slug");
			await invalidate("board:all");
		} finally {
			modBusy = false;
		}
	}

	async function deleteCatalogThread(threadId: string) {
		const err = await softDelete("thread", threadId);
		if (err) {
			toast.error(err);
			return;
		}
		await invalidate("board:slug");
	}

	async function restoreCatalogThread(threadId: string) {
		const err = await restoreDeleted("thread", threadId);
		if (err) {
			toast.error(err);
			return;
		}
		await invalidate("board:slug");
	}
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
				bannedMessage: (reason, until) =>
					until
						? $t("common.banned_until")
								.replace("{until}", formatDateTime(until))
								.replace("{reason}", reason)
						: $t("common.banned").replace("{reason}", reason),
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

			try {
				await followThread(thread.id, 0);
			} catch {
				toast.error($t("common.follow_failed"));
			}

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
	<div class="mt-8 mb-5 flex w-full flex-row flex-wrap items-center gap-3">
		<h3
			class={`scroll-m-20 text-2xl font-semibold tracking-tight ${isBoardDeleted ? "text-red-600" : ""}`}
		>
			{#if isBoardDeleted}
				{$t("common.deleted")}{" "}
			{/if}
			{data.board.name}
		</h3>
		{#if data.isAdmin}
			{#if isBoardDeleted}
				<Button
					variant="outline"
					size="sm"
					disabled={modBusy}
					onclick={() => void restoreBoard()}
				>
					<Update class="mr-1 h-4 w-4" />
					{$t("common.restore")}
				</Button>
			{:else}
				<Button
					variant="destructive"
					size="sm"
					disabled={modBusy}
					onclick={() => void deleteBoard()}
				>
					<Trash class="mr-1 h-4 w-4" />
					{$t("common.delete")}
				</Button>
			{/if}
		{/if}
	</div>
	{#if canCreateThread && !isBoardDeleted}
		<Button
			class="cursor-pointer"
			onclick={() => {
				isReplyOpen = !isReplyOpen;
			}}
		>
			{#if isReplyOpen}
				{$t("common.cancel")}
			{:else}
				{$t("common.threads.new")}
			{/if}
		</Button>
	{/if}
</div>
<Separator class="my-4" />

{#if canCreateThread && !isBoardDeleted && isReplyOpen}
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
	{#each data.board.threads as thread (thread.id)}
		<Card.Root class={thread.deleted_at ? "border-2 border-red-600" : ""}>
			<Card.Header>
				<Card.Title>
					<div class="flex flex-row justify-start items-center gap-3">
						{#if thread.deleted_at}
							<span class="text-red-600 font-semibold">{$t("common.deleted")}</span>
						{/if}
						{thread.title}
						{#if thread.is_author}
							<Badge>{$t("common.you")}</Badge>
						{/if}
						{#if data.isAdmin}
							{#if thread.deleted_at}
								<Button
									variant="outline"
									size="icon"
									title={$t("common.restore")}
									onclick={() => void restoreCatalogThread(thread.id)}
								>
									<Update />
								</Button>
							{:else}
								<Button
									variant="destructive"
									size="icon"
									title={$t("common.delete")}
									onclick={() => void deleteCatalogThread(thread.id)}
								>
									<Trash />
								</Button>
							{/if}
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
								type={thread.original_post.attachments[0].type ?? ""}
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
			<Card.Footer class="flex flex-row flex-wrap gap-2">
				<Button
					class="cursor-pointer"
					href={`/board/${data.slug}/thread/${thread.id}`}
				>
					{$t("common.reply")}
				</Button>
				<Button
					variant="outline"
					class="cursor-pointer"
					disabled={followBusyId === thread.id}
					onclick={() => void toggleFollowCatalog(thread)}
				>
					{#if followedIds.has(thread.id)}
						{$t("common.unfollow")}
					{:else}
						{$t("common.follow")}
					{/if}
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
