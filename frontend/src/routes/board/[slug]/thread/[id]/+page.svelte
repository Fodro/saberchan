<script lang="ts">
	import Separator from "$lib/components/ui/separator/separator.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import { formatDateTime, insertTagAtCursor } from "$lib/helpers.js";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import DoubleArrowDown from "svelte-radix/DoubleArrowDown.svelte";
	import { t } from "$lib/translations";
	import Draggable from "$lib/components/custom/Draggable.svelte";
	import { Textarea } from "$lib/components/ui/textarea/index.js";
	import { Label } from "$lib/components/ui/label/index.js";
	import { Checkbox } from "$lib/components/ui/checkbox/index.js";
	import { getContext, onMount } from "svelte";
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
	} from "svelte-radix";
	import { invalidate } from "$app/navigation";
	import PostBody from "$lib/components/custom/PostBody.svelte";
	import PostCard from "$lib/components/custom/PostCard.svelte";
	import { redirect } from "@sveltejs/kit";

	let newText = $state("");
	let newSage = $state(false);
	let newOP = $state(false);
	let isReplyOpen = $state(false);
	let counter: () => number = getContext("counter");

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
	<Draggable>
		<Card.Root class="w-[50vw] h-[50vh]">
			<Card.Header>
				<Card.Title>{$t("common.posts.new")}</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="grid grid-cols-1 w-full items-center gap-4">
					<div class="flex flex-row justify-start items-center gap-2">
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
					</div>
				</div>
			</Card.Content>
			<!-- TODO: add captcha -->
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
							await fetch("/api/post", {
								method: "POST",
								body: JSON.stringify({
									thread_id: data.thread.id,
									text: newText,
									sage: newSage,
									op_marker: newOP,
								}),
							});
							newText = "";
							newSage = false;
							newOP = false;
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
