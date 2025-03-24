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
	import { trackBoard } from "$lib/tracking";
	import type { Thread } from "$lib/types/thread.js";
	import { getContext, onMount } from "svelte";

	let { data } = $props();

	let counter: () => number = getContext("counter");

	let isReplyOpen = $state(false);

	let newTitle: string | null = $state(null);
	let newText: string | null = $state(null);

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
		on:click={() => {
			isReplyOpen = !isReplyOpen;
		}}
	>
		{#if isReplyOpen}
			Cancel
		{/if}
		{#if !isReplyOpen}
			New Thread
		{/if}
	</Button>
</div>
<Separator class="my-4" />

{#if isReplyOpen}
	<Draggable>
		<Card.Root class="w-[50vw] h-[50vh]">
			<Card.Header>
				<Card.Title>New Thread</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="grid grid-cols-1 w-full items-center gap-4">
					<div class="flex flex-col justify-start items-start gap-3">
						<Label>Title</Label>
						<Input
							placeholder="Type your title here"
							bind:value={newTitle}
						/>
					</div>
					<div class="flex flex-col justify-start items-start gap-3">
						<Label>Text</Label>
						<Textarea
							placeholder="Type your text here..."
							rows={10}
							class="min-h-[70%] w-full resize-none"
							bind:value={newText}
						/>
					</div>
				</div>
			</Card.Content>
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
						Cancel
					</Button>
					<Button
						on:click={async () => {
							const res = await fetch("/api/thread", {
								method: "POST",
								body: JSON.stringify({
									board_id: data.board.id,
									title: newTitle,
									original_post: {
										text: newText,
									},
								}),
							});
							const thread: Thread = await res.json();
							newText = null;
							newTitle = null;
							isReplyOpen = false;
							await window.open(
								`/board/${data.slug}/thread/${thread.id}`,
								"_blank",
								"noopener"
							);
						}}
					>
						Post!
					</Button>
				</div>
			</Card.Footer>
		</Card.Root>
	</Draggable>
{/if}

<div class="grid grid-cols-2 gap-4">
	{#each data.board.threads as thread}
		<Card.Root>
			<Card.Header>
				<Card.Title>
					<div class="flex flex-row justify-start items-center gap-3">
						{thread.title}
						{#if thread.is_author}
							<Badge>You</Badge>
						{/if}
					</div>
				</Card.Title>
				<Card.Description>
					anon #{thread.original_post.number}, replies: {thread.replies_count}
				</Card.Description>
			</Card.Header>
			<Card.Content>
				{#if thread.original_post.text.length <= 400}
					<p>{thread.original_post.text}</p>
				{/if}
				{#if thread.original_post.text.length > 400}
					<p>{thread.original_post.text.substring(0, 400)}...</p>
				{/if}
			</Card.Content>
			<Card.Footer>
				<Button
					href={`/board/${data.slug}/thread/${thread.id}`}
					target="_blank"
					rel="noreferrer noopener"
				>
					Reply
				</Button>
			</Card.Footer>
		</Card.Root>
	{/each}
</div>
