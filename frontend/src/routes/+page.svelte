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

	let newBoard: Board | undefined = $state(undefined);

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
						<Label for="name" class="text-right">Name</Label>
						<Input
							id="name"
							value="Pedro Duarte"
							class="col-span-3"
						/>
					</div>
					<div class="grid grid-cols-4 items-center gap-4">
						<Label for="username" class="text-right">Username</Label
						>
						<Input
							id="username"
							value="@peduarte"
							class="col-span-3"
						/>
					</div>
				</div>
				<Dialog.Footer>
					<Button type="submit">{$t("common.save")}</Button>
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
