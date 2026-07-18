<script lang="ts">
	import { t } from "$lib/translations";
	import Button from "../ui/button/button.svelte";
	import { Update } from "svelte-radix";
	import Input from "../ui/input/input.svelte";

	let isOpened: boolean = $state(false);
	let text: string = $state("");
	let src: string = $state("");
	let inflight: AbortController | undefined;
	let lastRefresh = $state(0);

	const {
		setCaptchaInput,
		setCaptchaToken,
		counter,
	}: {
		setCaptchaInput: (input: string) => void;
		setCaptchaToken: (token: string) => void;
		counter: number;
	} = $props();

	const fetchCaptcha = async () => {
		inflight?.abort();
		const ac = new AbortController();
		inflight = ac;

		try {
			const res = await fetch("/api/captcha", { signal: ac.signal });
			const blob = await res.blob();
			if (ac.signal.aborted) return;

			if (src) URL.revokeObjectURL(src);
			src = URL.createObjectURL(blob);
			setCaptchaToken(res.headers.get("x-captcha-token") ?? "");
			text = "";
			setCaptchaInput("");
		} catch (err) {
			if (err instanceof DOMException && err.name === "AbortError") return;
			throw err;
		}
	};

	const onInput = () => {
		setCaptchaInput(text.trim());
	};

	// Refresh only when parent bumps counter after a failed submit — not on mount.
	$effect(() => {
		if (counter > 0 && counter !== lastRefresh) {
			lastRefresh = counter;
			isOpened = true;
			void fetchCaptcha();
		}
	});
</script>

<div class="flex flex-row justify-start items-center gap-2">
	{#if !isOpened}
		<Button
			onclick={async () => {
				await fetchCaptcha();
				isOpened = true;
			}}
		>
			{$t("common.captcha.show")}
		</Button>
	{:else}
		<Button
			variant="outline"
			size="icon"
			onclick={async () => {
				await fetchCaptcha();
			}}
			class="flex-1"
		>
			<Update class="w-4 h-4" />
		</Button>
		<img {src} alt="captcha" />
		<Input
			placeholder={$t("common.captcha.placeholder")}
			class="flex-4"
			bind:value={text}
			oninput={onInput}
		/>
	{/if}
</div>
