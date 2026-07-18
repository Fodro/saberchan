<script lang="ts">
	import MediaLightbox from "$lib/components/custom/MediaLightbox.svelte";

	let {
		link,
		name,
	}: {
		link: string;
		name: string;
	} = $props();

	let open = $state(false);
</script>

<button
	type="button"
	class="flex cursor-pointer flex-col items-center justify-center gap-1 border-0 bg-transparent p-0 text-left"
	onclick={() => {
		open = true;
	}}
>
	<span class="text-sm text-primary underline-offset-4 hover:underline">{name}</span>
	<img src={link} alt={name} class="h-48 w-72 object-scale-down" />
</button>

<MediaLightbox bind:open {link} {name} />
