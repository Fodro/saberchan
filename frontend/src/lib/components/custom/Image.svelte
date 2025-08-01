<script lang="ts">
	import Button from "../ui/button/button.svelte";

	let {
		link,
		name,
	}: {
		link: string;
		name: string;
	} = $props();
</script>

<div class="flex flex-col justify-center items-center">
	<Button
		variant="link"
		class="cursor-pointer"
		href={`/view?link=${link}&name=${name}`}
		target="_blank"
		rel="noreferrer noopener"
	>
		{name}
	</Button>
	<img src={link} alt={name} class="h-48 w-72 object-scale-down" />
</div>
