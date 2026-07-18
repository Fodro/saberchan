<script lang="ts">
	import { invalidate } from "$app/navigation";
	import { formatDateTime } from "$lib/helpers";
	import { t } from "$lib/translations";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import { DoubleArrowDown, Trash, Update } from "svelte-radix";
	import { toast } from "svelte-sonner";
	import PostBody from "./PostBody.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import type { Post } from "$lib/types/post";
	import Image from "./Image.svelte";
	import { restoreDeleted, softDelete } from "$lib/adminModeration";

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

	let busy = $state(false);

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
					{/if}
					<Button variant="destructive">
						{$t("common.ban")}
					</Button>
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
			{#if post.attachments}
			<div class={`grid grid-cols-${colsCount} grid-rows-${rowsCount} items-center gap-2 flex-${imageFlex} p-2 border-r-7`}>
				{#each post.attachments as file, i (file.link ?? i)}
					<Image
						link={file.link ?? ""}
						name={file.name ?? ""}
						type={file.type ?? ""}
					/>
				{/each}
			</div>
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
