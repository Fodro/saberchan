<script lang="ts">
	import Separator from "$lib/components/ui/separator/separator.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import { formatDateTime } from "$lib/helpers.js";
	import Badge from "$lib/components/ui/badge/badge.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import DoubleArrowDown from "svelte-radix/DoubleArrowDown.svelte";

	let newText = $state("");

	$effect(() => {
		console.log(newText);
	});

	const { data } = $props();
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
	<Separator />
	<div class="grid grid-cols-1 gap-4">
		<Card.Root>
			<Card.Header>
				<Card.Title
					>anon #{data.thread.original_post.number} at {formatDateTime(
						data.thread.original_post.created_at,
					)}</Card.Title
				>
				<Card.Description>
					<Badge>OP</Badge>
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="flex flex-col justify-center items-start gap-2">
					<p class="leading-7">
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
						}}
					>
						Reply
					</Button>
				</div>
			</Card.Footer>
		</Card.Root>
		{#each data.thread.posts as post}
			<Card.Root>
				<Card.Header>
					<Card.Title
						>anon #{post.number} at {formatDateTime(
							post.created_at,
						)}</Card.Title
					>
					<Card.Description>
						{#if post.op_marker}
							<Badge>OP</Badge>
						{/if}
						{#if post.sage}
							<Badge variant="destructive">
								<DoubleArrowDown
									class="h-[1.2rem] w-[1.2rem]"
								/>
							</Badge>
						{/if}
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<div class="flex flex-col justify-center items-start gap-2">
						<p class="leading-7">
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
							}}
						>
							Reply
						</Button>
					</div>
				</Card.Footer>
			</Card.Root>
		{/each}
	</div>
</div>
