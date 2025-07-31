<script lang="ts">
	import { formatDateTime } from "$lib/helpers";
	import { t } from "$lib/translations";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import { DoubleArrowDown, Trash } from "svelte-radix";
	import { toast } from "svelte-sonner";
	import PostBody from "./PostBody.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import type { Post } from "$lib/types/post";
	import Image from "./Image.svelte";

	const {
		post,
		addToText,
		setReplyOpen,
		checkIsInText,
		isSigned,
	}: {
		post: Post;
		addToText: (txt: string) => void;
		setReplyOpen: (value: boolean) => void;
		checkIsInText: (txt: string) => boolean;
		isSigned: boolean;
	} = $props();
</script>

<Card.Root id={`${post.number}`} class="target:border-sky-500">
	<Card.Header>
		<Card.Title>
			<div class="flex flex-row justify-start items-center gap-2">
				{#if isSigned}
					<Button variant="destructive" size="icon">
						<Trash />
					</Button>
					<Button variant="destructive">
						{$t("common.ban")}
					</Button>
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
		<div class="flex flex-row justify-start items-start gap-2">
			<div class="grid auto-cols-auto auto-rows-auto items-center gap-2">
				{#each post.attachments as file}
					<Image link={file.link ?? ""} name={file.name ?? ""} />
				{/each}
			</div>
			<PostBody text={post.text} />
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
				on:click={async () => {
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
