<script lang="ts">
	import { invalidate } from "$app/navigation";
	import Separator from "$lib/components/ui/separator/separator.svelte";
	import type { File as FileType } from "$lib/types/attachment";
	import { Button } from "$lib/components/ui/button/index.js";
	import { t } from "$lib/translations";
	import { Label } from "$lib/components/ui/label/index.js";
	import { Checkbox } from "$lib/components/ui/checkbox/index.js";
	import { toast } from "svelte-sonner";
	import ComposeForm from "$lib/components/custom/ComposeForm.svelte";
	import PostCard from "$lib/components/custom/PostCard.svelte";
	import { buildAttachments, composeErrorMessageFactory, submitCompose } from "$lib/compose";

	let newText = $state("");
	let newSage = $state(false);
	let newOP = $state(false);
	let isReplyOpen = $state(false);
	let submitting = $state(false);
	let captchaInput = $state("");
	let captchaToken = $state("");
	let captchaCounter = $state(0);
	let filesList: FileType[] = $state([]);

	let formX = $state(0);
	let formY = $state(0);
	let formPinned = $state(true);

	const { data } = $props();

	const composeErrorMessage = composeErrorMessageFactory((key) => $t(key));

	const checkIsInText = (txt: string): boolean => {
		return newText.includes(txt);
	};

	const addToText = (txt: string) => {
		newText += txt;
	};

	const setReplyOpen = (value: boolean) => {
		isReplyOpen = value;
	};

	const submitNewPost = async () => {
		submitting = true;
		try {
			const attachments = buildAttachments(filesList);
			const payload = {
				thread_id: data.thread.id,
				text: newText,
				sage: newSage,
				op_marker: newOP,
				attachments,
				captcha: {
					input: captchaInput,
					token: captchaToken,
				},
			};

			const result = await submitCompose({
				endpoint: "/api/post",
				payload,
				text: newText,
				captchaInput,
				captchaToken,
				files: filesList,
				errorMessage: composeErrorMessage,
				captchaFailedMessage: $t("common.captcha.failed"),
			});

			if (!result.ok) {
				toast.error(result.message);
				if (result.bumpCaptcha) captchaCounter += 1;
				return;
			}

			newText = "";
			newSage = false;
			newOP = false;
			isReplyOpen = false;
			filesList = [];

			await invalidate("thread:id");
		} finally {
			submitting = false;
		}
	};
</script>

<svelte:head>
	<title>
		/{data.slug}/ - #{data.thread.original_post.number} - {data.thread.original_post.text.substring(
			0,
			15,
		)}...
	</title>
</svelte:head>

<svelte:window bind:scrollX={formX} bind:scrollY={formY} />

<div class="flex flex-col justify-center items-start gap-2">
	<h3 class="mt-8 scroll-m-20 text-2xl font-semibold tracking-tight mb-5">
		{data.thread.title}
	</h3>
	<Button
		class="cursor-pointer"
		onclick={() => {
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
	<ComposeForm
		heading={$t("common.posts.new")}
		textareaId="new-post-area"
		initialLeft={formX}
		initialTop={formY}
		bind:formPinned
		bind:text={newText}
		bind:files={filesList}
		bind:captchaInput
		bind:captchaToken
		{captchaCounter}
		onCancel={() => {
			isReplyOpen = false;
		}}
		onSubmit={submitNewPost}
		{submitting}
	>
		{#snippet afterToolbar()}
			<div class="flex md:flex-row flex-col justify-start md:items-center items-start gap-2">
				<div class="flex flex-row justify-start items-center gap-2">
					<Checkbox id="sage" bind:checked={newSage} class="cursor-pointer" />
					<Label>
						{$t("common.fields.sage")}
					</Label>
				</div>
				{#if data.thread.original_post.is_author}
					<div class="flex flex-row justify-start items-center gap-2">
						<Checkbox id="op" bind:checked={newOP} class="cursor-pointer" />
						<Label>
							{$t("common.op")}
						</Label>
					</div>
				{/if}
			</div>
		{/snippet}
	</ComposeForm>
{/if}
