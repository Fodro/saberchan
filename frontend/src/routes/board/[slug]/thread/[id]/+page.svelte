<script lang="ts">
	const { data } = $props();
</script>

<svelte:head>
	<title>
		/{data.slug}/ - #{data.thread.original_post.number} - {data.thread.original_post.text.substring(
			0,
			15,
		)}...
	</title>
</svelte:head>

<p>test</p>
