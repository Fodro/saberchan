<script lang="ts">
	import { t } from "$lib/translations";
	import Button from "../ui/button/button.svelte";
	import { Update } from "svelte-radix";
	import Input from "../ui/input/input.svelte";

	let isOpened: boolean = $state(false);
	let text: string = $state("");
	let src: string = $state("");

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
		const res = await fetch("/api/captcha");
		const blob = await res.blob();
		src = URL.createObjectURL(blob);
		setCaptchaToken(res.headers.get("x-captcha-token") ?? "");
	};

	$effect(() => {
		setCaptchaInput(text);
	});

	$effect(() => {
		counter;
		fetchCaptcha();
		text = "";
	})
</script>

<div class="flex flex-row justify-start items-center gap-2">
	{#if !isOpened}
		<Button
			on:click={async () => {
				await fetchCaptcha();
				isOpened = true;
				text = "";
			}}
		>
			{$t("common.captcha.show")}
		</Button>
	{:else}
		<Button
			variant="outline"
			size="icon"
			on:click={async () => {
				await fetchCaptcha();
				text = "";
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
		/>
	{/if}
</div>
