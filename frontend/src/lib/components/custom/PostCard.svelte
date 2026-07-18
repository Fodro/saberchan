<script lang="ts">
	import { invalidate } from "$app/navigation";
	import { formatDateTime } from "$lib/helpers";
	import { t } from "$lib/translations";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import { Button, buttonVariants } from "$lib/components/ui/button/index.js";
	import { DoubleArrowDown, Trash, Update } from "svelte-radix";
	import { toast } from "svelte-sonner";
	import PostBody from "./PostBody.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import * as Dialog from "$lib/components/ui/dialog/index.js";
	import Label from "$lib/components/ui/label/label.svelte";
	import Input from "$lib/components/ui/input/input.svelte";
	import type { Post } from "$lib/types/post";
	import AttachmentGallery from "./AttachmentGallery.svelte";
	import {
		banPost,
		restoreDeleted,
		softDelete,
		type BanDuration,
	} from "$lib/adminModeration";
	import { cn } from "$lib/utils.js";

	const {
		post,
		addToText,
		setReplyOpen,
		checkIsInText,
		isAdmin = false,
	}: {
		post: Post;
		addToText: (txt: string) => void;
		setReplyOpen: (value: boolean) => void;
		checkIsInText: (txt: string) => boolean;
		isAdmin?: boolean;
	} = $props();

	const colsCount = $derived(
		!post.attachments ? 0 : post.attachments.length < 2 ? 1 : 2,
	);
	const rowsCount = $derived(
		!post.attachments ? 0 : post.attachments.length < 3 ? 1 : 2,
	);
	const imageFlex = $derived(!post.attachments ? 0 : 1);
	const isDeleted = $derived(Boolean(post.deleted_at));

	const banDurations: BanDuration[] = ["1h", "1d", "7d", "30d", "permanent"];

	let busy = $state(false);
	let banOpen = $state(false);
	let banReason = $state("");
	let banDuration: BanDuration = $state("1d");

	async function onDelete() {
		busy = true;
		try {
			const err = await softDelete("post", post.id);
			if (err) {
				toast.error(err);
				return;
			}
			await invalidate("thread:id");
		} finally {
			busy = false;
		}
	}

	async function onRestore() {
		busy = true;
		try {
			const err = await restoreDeleted("post", post.id);
			if (err) {
				toast.error(err);
				return;
			}
			await invalidate("thread:id");
		} finally {
			busy = false;
		}
	}

	async function onBan() {
		const reason = banReason.trim();
		if (!reason) {
			toast.error($t("common.ban_reason"));
			return;
		}
		busy = true;
		try {
			const err = await banPost(post.id, reason, banDuration);
			if (err) {
				toast.error(err);
				return;
			}
			toast.success($t("common.ban_success"));
			banOpen = false;
			banReason = "";
			banDuration = "1d";
			await invalidate("thread:id");
		} finally {
			busy = false;
		}
	}
</script>

<Card.Root
	id={`${post.number}`}
	class={`target:border-sky-500 ${isDeleted ? "border-2 border-red-600" : ""}`}
>
	<Card.Header>
		<Card.Title>
			<div class="flex flex-row justify-start items-center gap-2">
				{#if isAdmin}
					{#if isDeleted}
						<Button
							variant="outline"
							size="icon"
							disabled={busy}
							title={$t("common.restore")}
							onclick={() => void onRestore()}
						>
							<Update />
						</Button>
					{:else}
						<Button
							variant="destructive"
							size="icon"
							disabled={busy}
							title={$t("common.delete")}
							onclick={() => void onDelete()}
						>
							<Trash />
						</Button>
						<Dialog.Root bind:open={banOpen}>
							<Dialog.Trigger
								class={cn(
									buttonVariants({ variant: "destructive" }),
									"cursor-pointer",
								)}
								disabled={busy}
							>
								{$t("common.ban")}
							</Dialog.Trigger>
							<Dialog.Content class="sm:max-w-[425px]">
								<Dialog.Header>
									<Dialog.Title>{$t("common.ban")}</Dialog.Title>
									<Dialog.Description>
										anon #{post.number}
									</Dialog.Description>
								</Dialog.Header>
								<div class="grid gap-4 py-2">
									<div class="grid gap-2">
										<Label for={`ban-reason-${post.id}`}>{$t("common.ban_reason")}</Label>
										<Input
											id={`ban-reason-${post.id}`}
											bind:value={banReason}
										/>
									</div>
									<div class="grid gap-2">
										<Label for={`ban-duration-${post.id}`}>{$t("common.ban_duration")}</Label>
										<select
											id={`ban-duration-${post.id}`}
											class="border-input bg-background ring-offset-background focus-visible:ring-ring flex h-9 w-full rounded-md border px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
											bind:value={banDuration}
										>
											{#each banDurations as d (d)}
												<option value={d}>{$t(`common.ban_presets.${d}`)}</option>
											{/each}
										</select>
									</div>
								</div>
								<Dialog.Footer>
									<Button disabled={busy} onclick={() => void onBan()}>
										{$t("common.ban_submit")}
									</Button>
								</Dialog.Footer>
							</Dialog.Content>
						</Dialog.Root>
					{/if}
				{/if}
				{#if isDeleted}
					<span class="text-red-600 font-semibold">{$t("common.deleted")}</span>
				{/if}
				anon #{post.number}
				{$t("common.posts.at")}
				{formatDateTime(post.created_at)}
				{#if post.is_author}
					<Badge>{$t("common.you")}</Badge>
				{/if}
				{#if post.op_marker}
					<Badge>{$t("common.op")}</Badge>
				{/if}
				{#if post.sage}
					<Badge variant="destructive">
						<DoubleArrowDown class="h-[1.2rem] w-[1.2rem]" />
					</Badge>
				{/if}
			</div>
		</Card.Title>
	</Card.Header>
	<Card.Content>
		<div class="flex flex-row justify-start items-start gap-3">
			{#if post.attachments && post.attachments.length > 0}
				<AttachmentGallery
					attachments={post.attachments}
					{colsCount}
					{rowsCount}
					{imageFlex}
				/>
			{/if}
			<PostBody text={post.text} additionalClass="" />
		</div>
	</Card.Content>
	<Card.Footer>
		<div
			class="flex flex-row justify-start items-center gap-4 w-full h-full"
		>
			<Button
				class="cursor-pointer"
				variant="secondary"
				onclick={() => {
					const toAppend = `>>${post.number}\n`;
					if (!checkIsInText(toAppend)) {
						addToText(toAppend);
					}
					setReplyOpen(true);
				}}
			>
				{$t("common.reply")}
			</Button>
			<Button
				class="cursor-pointer"
				variant="outline"
				onclick={async () => {
					const base = window.location.href.split("#");
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
