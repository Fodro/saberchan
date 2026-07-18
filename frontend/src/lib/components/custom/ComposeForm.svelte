<script lang="ts">
	import type { Snippet } from "svelte";
	import Draggable from "$lib/components/custom/Draggable.svelte";
	import Captcha from "$lib/components/custom/Captcha.svelte";
	import FileUploader from "$lib/components/custom/FileUploader.svelte";
	import MarkupToolbar from "$lib/components/custom/MarkupToolbar.svelte";
	import Button from "$lib/components/ui/button/button.svelte";
	import * as Card from "$lib/components/ui/card/index.js";
	import Label from "$lib/components/ui/label/label.svelte";
	import { Textarea } from "$lib/components/ui/textarea/index.js";
	import type { File as FileType } from "$lib/types/attachment";
	import { t } from "$lib/translations";
	import { DrawingPin, DrawingPinFilled } from "svelte-radix";

	let {
		heading,
		textareaId,
		initialLeft,
		initialTop,
		formPinned = $bindable(true),
		text = $bindable(""),
		files = $bindable([] as FileType[]),
		captchaInput = $bindable(""),
		captchaToken = $bindable(""),
		captchaCounter,
		beforeText,
		afterToolbar,
		onCancel,
		onSubmit,
		submitting = false,
	}: {
		heading: string;
		textareaId: string;
		initialLeft: number;
		initialTop: number;
		formPinned?: boolean;
		text?: string;
		files?: FileType[];
		captchaInput?: string;
		captchaToken?: string;
		captchaCounter: number;
		beforeText?: Snippet;
		afterToolbar?: Snippet;
		onCancel: () => void;
		onSubmit: () => void | Promise<void>;
		submitting?: boolean;
	} = $props();

	let draggable: { preparePinChange: (nextPinned: boolean) => void } | undefined =
		$state();

	function togglePinned() {
		const next = !formPinned;
		draggable?.preparePinChange(next);
		formPinned = next;
	}
</script>

<Draggable bind:this={draggable} {initialLeft} {initialTop} pinned={formPinned}>
	<Card.Root class="w-[100%] h-[100%]">
		<Card.Header>
			<Card.Title>
				<div class="flex flex-row items-center h-[5%]">
					<div class="flex flex-row flex-1">
						<p class="text-muted-foreground">{heading}</p>
					</div>
					<div class="flex flex-row-reverse flex-1">
						<Button
							class="cursor-pointer md:flex hidden"
							variant="outline"
							size="icon"
							onclick={togglePinned}
						>
							{#if formPinned}
								<DrawingPinFilled />
							{:else}
								<DrawingPin />
							{/if}
						</Button>
					</div>
				</div>
			</Card.Title>
			<Card.Description>
				<p class="text-muted-foreground md:flex hidden">
					{$t("common.draggable")}
				</p>
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="grid grid-cols-1 w-full items-center gap-2">
				<div class="flex flex-col justify-start items-start gap-2">
					{#if beforeText}
						{@render beforeText()}
					{/if}
					<Label>{$t("common.fields.text")}</Label>
					<MarkupToolbar {textareaId} />
					{#if afterToolbar}
						{@render afterToolbar()}
					{/if}
					<Textarea
						id={textareaId}
						placeholder={$t("common.fields.text_placeholder")}
						rows={10}
						class="min-h-[70%] w-full resize-none"
						bind:value={text}
					/>
					<FileUploader bind:value={files} />
					<Captcha bind:captchaInput bind:captchaToken counter={captchaCounter} />
				</div>
			</div>
		</Card.Content>
		<Card.Footer>
			<div class="flex flex-row justify-start items-center gap-4 w-full h-full">
				<Button class="cursor-pointer" variant="secondary" onclick={onCancel}>
					{$t("common.cancel")}
				</Button>
				<Button class="cursor-pointer" disabled={submitting} onclick={() => void onSubmit()}>
					{$t("common.post")}
				</Button>
			</div>
		</Card.Footer>
	</Card.Root>
</Draggable>
