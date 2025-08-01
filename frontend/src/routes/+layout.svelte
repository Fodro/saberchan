<script lang="ts">
	import "../app.css";
	import { Toaster } from "$lib/components/ui/sonner";
	import * as Menubar from "$lib/components/ui/menubar/index.js";
	import { ModeWatcher } from "mode-watcher";
	import Sun from "svelte-radix/Sun.svelte";
	import Moon from "svelte-radix/Moon.svelte";
	import Home from "svelte-radix/Home.svelte";
	import { toggleMode } from "mode-watcher";
	import { Button } from "$lib/components/ui/button/index.js";
	import { goto, invalidate, invalidateAll } from "$app/navigation";
	import { onMount, setContext } from "svelte";
	import { t } from "$lib/translations";
	import { CounterClockwiseClock, Update } from "svelte-radix";

	let intervals = [1, 5, 10, 15, 30, 60];

	let counter = $state(0);
	let interval: number | undefined = $state(9999);

	let localeLoading = $state(false);

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
<div class="flex flex-col p-4 h-screen gap-2">
	<div class="flex flex-row justify-start items-center">
		<div class="flex flex-row basis-1/2 gap-3 justify-start items-center pb-4">
			<h2
				class="mt-10 scroll-m-20 text-3xl font-semibold tracking-tight transition-colors first:mt-0"
			>
				Saberchan
			</h2>
			<Button
				class="cursor-pointer"
				on:click={() => {
					goto("/");
				}}
				size="icon"
				variant="outline"
			>
				<Home class="absolute h-[1.2rem] w-[1.2rem]" />
			</Button>
		</div>
		<div
			class="flex flex-row-reverse basis-1/2 gap-3 justify-start items-center"
		>
			<!-- TODO: make a select -->
			{#if data.i18n.locale === "ru"}
				<Button
					variant="secondary"
					class="cursor-pointer"
					on:click={async () => {
						localeLoading = true;
						await fetch("/locale", {
							method: "POST",
							body: JSON.stringify({
								locale: "en",
							}),
						});
						localeLoading = false;
						invalidateAll();
					}}
				>
					{#if localeLoading}
						<CounterClockwiseClock />
					{:else}
						RU
					{/if}
				</Button>
			{/if}
			{#if data.i18n.locale === "en"}
				<Button
					variant="secondary"
					class="cursor-pointer"
					on:click={async () => {
						localeLoading = true;
						await fetch("/locale", {
							method: "POST",
							body: JSON.stringify({
								locale: "ru",
							}),
						});
						localeLoading = false;
						invalidateAll();
					}}
				>
					{#if localeLoading}
						<CounterClockwiseClock />
					{:else}
						EN
					{/if}
				</Button>
			{/if}
			<Button on:click={toggleMode} size="icon" class="cursor-pointer">
				<Sun
					class="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0"
				/>
				<Moon
					class="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100"
				/>
			</Button>
			{#if !data.signed}
				<Button href={data.loginUrl}>{$t("common.login")}</Button>
			{:else}
				<Button variant="destructive" href="/admin/auth/signOut">
					[BETA] {$t("common.clearTokens")}
				</Button>
				<Button
					href={data.idToken
						? `${data.logoutUrl}&id_token_hint=${data.idToken}`
						: data.logoutUrl}
				>
					{$t("common.logout")}
				</Button>
				<p class="text-muted-foreground">
					[ADMIN] {$t("common.logged_in")}
					{data.username}
				</p>
			{/if}
		</div>
	</div>
	<Menubar.Root>
		<Menubar.Menu>
			<Menubar.Trigger>{$t("common.boards.list_title")}</Menubar.Trigger>
			<Menubar.Content>
				<Menubar.Sub>
					<Menubar.SubTrigger
						>{$t("common.recent")}</Menubar.SubTrigger
					>
					<Menubar.SubContent>
						{#if data.meta.recentBoards.length === 0}
							<Menubar.Item disabled>
								{$t("common.boards.noRecent")}
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
				{$t("common.autoreload")}: {interval && interval !== 9999
					? `${interval}s`
					: $t("common.disabled")}
			</Menubar.Trigger>
			<Menubar.Content>
				<Menubar.Item
					inset
					on:click={() => {
						interval = undefined;
					}}
				>
					{$t("common.disable")}
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
		<Button
			variant="ghost"
			size="icon"
			class="cursor-pointer"
			on:click={() => {
				invalidateAll();
			}}
		>
			<Update class="w-4 h-4" />
		</Button>
	</Menubar.Root>

	{@render children()}
</div>

<Toaster />
