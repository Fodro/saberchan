<script lang="ts">
	import "../app.css";
	import * as Menubar from "$lib/components/ui/menubar/index.js";
	import { ModeWatcher } from "mode-watcher";
	import Sun from "svelte-radix/Sun.svelte";
	import Moon from "svelte-radix/Moon.svelte";
	import Home from "svelte-radix/Home.svelte";
	import { toggleMode } from "mode-watcher";
	import { Button } from "$lib/components/ui/button/index.js";
	import { goto, invalidate } from "$app/navigation";
	import { onMount, setContext } from "svelte";

	let intervals = [1, 5, 10, 15, 30, 60];

	let counter = $state(0);
	let interval: number | undefined = $state(9999);

	setContext("counter", () => {
		return counter;
	});

	$effect(() => {
		if (interval !== 9999) {
			localStorage.setItem("autoreload", `${interval}`);
		}
	});

	$effect(() => {
		if (interval && interval !== 9999) {
			const id = setInterval(() => {
				counter += 1;
			}, interval * 1000);

			return () => {
				clearInterval(id);
			};
		}
	});

	$effect(() => {
		counter;
		invalidate("board:all");
	});

	onMount(() => {
		const savedInterval = localStorage.getItem("autoreload");
		interval = savedInterval !== null ? +savedInterval : 9999;
	});

	let { children, data } = $props();
</script>

<ModeWatcher />
<div class="p-4 h-screen">
	<div class="flex flex-row">
		<div class="flex flex-row basis-1/2 gap-3">
			<h2
				class="mt-10 scroll-m-20 border-b pb-2 text-3xl font-semibold tracking-tight transition-colors first:mt-0"
			>
				Saberchan
			</h2>
			<Button
				on:click={() => {
					goto("/");
				}}
				size="icon"
				variant="outline"
			>
				<Home class="absolute h-[1.2rem] w-[1.2rem]" />
			</Button>
		</div>
		<div class="flex flex-row-reverse basis-1/2 gap-3">
			<Button on:click={toggleMode} size="icon">
				<Sun
					class="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0"
				/>
				<Moon
					class="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100"
				/>
			</Button>
		</div>
	</div>
	<Menubar.Root>
		<Menubar.Menu>
			<Menubar.Trigger>Boards</Menubar.Trigger>
			<Menubar.Content>
				<Menubar.Sub>
					<Menubar.SubTrigger>Recent</Menubar.SubTrigger>
					<Menubar.SubContent>
						{#if data.meta.recentBoards.length === 0}
							<Menubar.Item disabled>
								No recent boards
							</Menubar.Item>
						{/if}
						{#each data.meta.recentBoards as alias}
							<Menubar.Item
								href={`/board/${alias}`}
								target="_blank"
								rel="noreferrer noopener"
							>
								/{alias}/
							</Menubar.Item>
						{/each}
					</Menubar.SubContent>
				</Menubar.Sub>
				<Menubar.Separator />
				<Menubar.Sub>
					<Menubar.SubTrigger>All</Menubar.SubTrigger>
					<Menubar.SubContent>
						{#each data.boards as board}
							<Menubar.Item
								href={`/board/${board.alias}`}
								target="_blank"
								rel="noreferrer noopener"
							>
								/{board.alias}/ - {board.name}
							</Menubar.Item>
						{/each}
					</Menubar.SubContent>
				</Menubar.Sub>
			</Menubar.Content>
		</Menubar.Menu>
		<Menubar.Menu>
			<Menubar.Trigger>
				Autoreload: {interval && interval !== 9999
					? `${interval}s`
					: "disabled"}
			</Menubar.Trigger>
			<Menubar.Content>
				<Menubar.Item
					inset
					on:click={() => {
						interval = undefined;
					}}
				>
					Disable
				</Menubar.Item>
				{#each intervals as i}
					<Menubar.Item
						inset
						on:click={() => {
							interval = i;
						}}
					>
						{`${i}s`}
					</Menubar.Item>
				{/each}
			</Menubar.Content>
		</Menubar.Menu>
	</Menubar.Root>

	{@render children()}
</div>
