<script lang="ts">
	import { HoverCard, HoverCardTrigger } from "$lib/components/ui/hover-card";
	import { Separator } from "$lib/components/ui/separator/index.js";
	import HoverCardContent from "$lib/components/ui/hover-card/hover-card-content.svelte";
	import { t } from "$lib/translations";
	import Button from "$lib/components/ui/button/button.svelte";
	import type { Board } from "$lib/types/board.js";
	import * as Dialog from "$lib/components/ui/dialog/index.js";
	import { buttonVariants } from "$lib/components/ui/button/index.js";
	import { Label } from "$lib/components/ui/label/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import Checkbox from "$lib/components/ui/checkbox/checkbox.svelte";
	import { toast } from "svelte-sonner";

	let newAlias: string | undefined = $state(undefined);
	let newName: string | undefined = $state(undefined);
	let newDescription: string | undefined = $state(undefined);
	let newLocked: boolean = $state(false);

	let { data } = $props();
</script>

<svelte:head>
	<title>Saberchan - {$t("common.boards.list_title")}</title>
</svelte:head>

<div class="flex flex-row justify-start items-center gap-2">
	<h3 class="mt-8 scroll-m-20 text-2xl font-semibold tracking-tight mb-5">
		{$t("common.boards.list_title")}
	</h3>
	{#if data.signed}
		<Dialog.Root>
			<Dialog.Trigger class={buttonVariants({ variant: "default" })}>
				{$t("common.admin.add_board")}
			</Dialog.Trigger>
			<Dialog.Content class="sm:max-w-[425px]">
				<Dialog.Header>
					<Dialog.Title>{$t("common.admin.add_board")}</Dialog.Title>
				</Dialog.Header>
				<div class="grid gap-4 py-4">
					<div class="grid grid-cols-4 items-center gap-4">
						<Label for="name" class="text-right"
							>{$t("common.boards.name")}</Label
						>
						<Input
							id="name"
							class="col-span-3"
							bind:value={newName}
						/>
					</div>
					<div class="grid grid-cols-4 items-center gap-4">
						<Label for="alias" class="text-right"
							>{$t("common.boards.alias")}</Label
						>
						<Input
							id="alias"
							class="col-span-3"
							bind:value={newAlias}
						/>
					</div>
					<div class="grid grid-cols-4 items-center gap-4">
						<Label for="description" class="text-right"
							>{$t("common.boards.description")}</Label
						>
						<Input
							id="description"
							class="col-span-3"
							bind:value={newDescription}
						/>
					</div>
					<div class="grid grid-cols-4 items-center gap-4">
						<Label for="locked" class="text-center col-span-3"
							>{$t("common.boards.locked")}</Label
						>
						<Checkbox
							id="locked"
							class="col-span-1"
							bind:checked={newLocked}
						/>
					</div>
				</div>
				<Dialog.Footer>
					<div
						class="flex flex-row justify-start items-center gap-4 w-full h-full"
					>
						<Dialog.Close asChild>
							<Button
								class="cursor-pointer"
								on:click={async () => {
									if (
										newName &&
										newAlias &&
										newDescription &&
										data.username
									) {
										const board: Board = {
											id: "",
											name: newName,
											alias: newAlias,
											description: newDescription,
											locked: newLocked,
											author: data.username,
											threads: [],
										};
										const res = await fetch("/api/board", {
											method: "POST",
											body: JSON.stringify(board),
										});
										if (
											res.status != 201 &&
											res.status != 200
										) {
											toast.error(await res.text());
											return;
										}
										await window.open(
											"/",
											"_self",
											"noopener",
										);
									}
								}}
							>
								{$t("common.save")}
							</Button>
						</Dialog.Close>
					</div>
				</Dialog.Footer>
			</Dialog.Content>
		</Dialog.Root>
	{/if}
</div>

<Separator class="my-4" />

<div class="flex flex-col gap-3 justify-start items-start h-auto">
	{#each data.boards as board}
		<div class="flex flex-row justify-start items-center">
			<HoverCard>
				<HoverCardTrigger
					href={`/board/${board.alias}`}
					target="_blank"
					rel="noreferrer noopener"
					class="rounded-sm underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-8 focus-visible:outline-black"
				>
					/{board.alias}/ - {board.name}
				</HoverCardTrigger>
				<HoverCardContent>
					{board.description}
				</HoverCardContent>
			</HoverCard>
		</div>
	{/each}
</div>
