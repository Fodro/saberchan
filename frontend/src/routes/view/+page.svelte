<script lang="ts">
	const { data } = $props();
</script>

<div class="flex flex-col justify-center items-center w-auto h-auto">
	<h1>{data.imgName}</h1>
	<img class="object-scale-down w-300 h-150" src={data.imgLink} alt={data.imgName} />
</div>
