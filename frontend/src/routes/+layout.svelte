<script lang="ts">
	import "../app.css";
	import { Toaster } from "$lib/components/ui/sonner";
	import * as Menubar from "$lib/components/ui/menubar/index.js";
	import { ModeWatcher } from "mode-watcher";
	import Sun from "svelte-radix/Sun.svelte";
	import Moon from "svelte-radix/Moon.svelte";
	import Home from "svelte-radix/Home.svelte";
	import { toggleMode } from "mode-watcher";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import { goto, invalidate, invalidateAll } from "$app/navigation";
	import { page } from "$app/state";
	import { onMount } from "svelte";
	import { unfollowThread } from "$lib/followed";
	import { t } from "$lib/translations";
	import LocaleSwitcher from "$lib/components/custom/LocaleSwitcher.svelte";
	import { Cross2, Update } from "svelte-radix";
	import { toast } from "svelte-sonner";

	// Prefer longer intervals — 1–10s hammers the BFF/backend.
	let intervals = [15, 30, 60];

	let interval: number | undefined = $state(9999);

	$effect(() => {
		if (interval !== 9999) {
			localStorage.setItem("autoreload", `${interval}`);
		}
	});

	$effect(() => {
		if (!interval || interval === 9999) return;

		const tick = () => {
			if (typeof document !== "undefined" && document.visibilityState !== "visible") {
				return;
			}
			const path = page.url.pathname;
			if (path.includes("/thread/")) {
				void invalidate("thread:id");
			} else if (path.startsWith("/board/")) {
				void invalidate("board:slug");
			}
			void invalidate("followed:list");
			// Do not invalidate board:all on the timer — nav list is not the live surface.
		};

		const id = setInterval(tick, interval * 1000);
		const onVisibility = () => {
			if (document.visibilityState === "visible") tick();
		};
		document.addEventListener("visibilitychange", onVisibility);

		return () => {
			clearInterval(id);
			document.removeEventListener("visibilitychange", onVisibility);
		};
	});

	onMount(() => {
		const savedInterval = localStorage.getItem("autoreload");
		const parsed = savedInterval !== null ? +savedInterval : 9999;
		// Migrate aggressive saved prefs to at least 15s (or off).
		if (parsed !== 9999 && parsed > 0 && parsed < 15) {
			interval = 30;
		} else {
			interval = parsed;
		}
	});

	let { children, data } = $props();

	async function unfollowFromMenu(id: string) {
		try {
			await unfollowThread(id);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : String(e));
		}
	}
</script>

<ModeWatcher />
<div class="flex flex-col p-4 h-screen gap-2">
	<div class="flex md:flex-row flex-col justify-start items-center">
		<div class="flex flex-row basis-1/2 gap-3 justify-start items-center pb-4">
			<h2
				class="mt-10 scroll-m-20 text-3xl font-semibold tracking-tight transition-colors first:mt-0"
			>
				Saberchan
			</h2>
			<Button
				class="cursor-pointer"
				onclick={() => {
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
			<LocaleSwitcher current={data.i18n.locale} />
			<Button onclick={toggleMode} size="icon" class="cursor-pointer">
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
					onclick={() => {
						interval = undefined;
					}}
				>
					{$t("common.disable")}
				</Menubar.Item>
				{#each intervals as i}
					<Menubar.Item
						inset
						onclick={() => {
							interval = i;
						}}
					>
						{`${i}s`}
					</Menubar.Item>
				{/each}
			</Menubar.Content>
		</Menubar.Menu>
		<Menubar.Menu>
			<Menubar.Trigger class="gap-2">
				{$t("common.followed_threads")}
				{#if data.followed.dirty_count > 0}
					<Badge variant="secondary" class="ml-1">
						{data.followed.dirty_count}
					</Badge>
				{/if}
			</Menubar.Trigger>
			<Menubar.Content class="min-w-56">
				{#if data.followed.threads.length === 0}
					<Menubar.Item disabled>
						{$t("common.followed_empty")}
					</Menubar.Item>
				{:else}
					{#each data.followed.threads as thread (thread.id)}
						<Menubar.Item class="gap-2 p-0 focus:bg-transparent data-[highlighted]:bg-transparent">
							<a
								href={thread.href}
								target="_blank"
								rel="noreferrer noopener"
								class="hover:bg-accent flex min-w-0 flex-1 items-center justify-between gap-2 rounded-sm px-2 py-1.5"
							>
								<span class="truncate max-w-40">{thread.title}</span>
								{#if thread.dead}
									<span class="text-muted-foreground shrink-0"
										>{$t("common.followed_dead")}</span
									>
								{:else if thread.new_posts > 0}
									<Badge variant="secondary" class="shrink-0"
										>{thread.new_posts}</Badge
									>
								{/if}
							</a>
							<Button
								type="button"
								variant="ghost"
								size="icon"
								class="text-muted-foreground hover:text-destructive mr-1 h-7 w-7 shrink-0 cursor-pointer"
								title={$t("common.unfollow")}
								onclick={(e) => {
									e.preventDefault();
									e.stopPropagation();
									void unfollowFromMenu(thread.id);
								}}
							>
								<Cross2 class="h-3.5 w-3.5" />
							</Button>
						</Menubar.Item>
					{/each}
				{/if}
			</Menubar.Content>
		</Menubar.Menu>
		<Button
			variant="ghost"
			size="icon"
			class="cursor-pointer"
			onclick={() => {
				invalidateAll();
			}}
		>
			<Update class="w-4 h-4" />
		</Button>
	</Menubar.Root>

	{@render children()}
</div>

<Toaster />
