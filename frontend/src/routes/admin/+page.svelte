<script lang="ts">
	import Button from "$lib/components/ui/button/button.svelte";

	let { data } = $props();
</script>

<div class="flex flex-row justify-start items-center gap-2 pt-5">
	{#if !data.signed}
		<Button href={data.loginUrl}>Login</Button>
	{:else}
		<Button href={`${data.logoutUrl}&id_token_hint=${data.idToken}`}>Logout</Button>
	{/if}
</div>
