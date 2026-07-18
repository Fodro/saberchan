<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { APP_LOCALES, localeFlag, type AppLocale } from '$lib/locales';
	import { invalidateAll } from '$app/navigation';
	import { CounterClockwiseClock } from 'svelte-radix';

	let {
		current,
	}: {
		current: string;
	} = $props();

	let open = $state(false);
	let loading = $state(false);
	let rootEl: HTMLDivElement | undefined = $state();

	const currentFlag = $derived(localeFlag(current));

	async function choose(code: AppLocale) {
		if (code === current || loading) {
			open = false;
			return;
		}
		loading = true;
		open = false;
		try {
			await fetch('/locale', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ locale: code }),
			});
			await invalidateAll();
		} finally {
			loading = false;
		}
	}

	function onDocPointerDown(e: PointerEvent) {
		if (!open || !rootEl) return;
		if (e.target instanceof Node && !rootEl.contains(e.target)) {
			open = false;
		}
	}
</script>

<svelte:document onpointerdown={onDocPointerDown} />

<div class="relative" bind:this={rootEl}>
	<Button
		type="button"
		variant="secondary"
		size="icon"
		class="cursor-pointer text-base"
		aria-haspopup="listbox"
		aria-expanded={open}
		aria-label="Language"
		disabled={loading}
		onclick={() => {
			open = !open;
		}}
	>
		{#if loading}
			<CounterClockwiseClock class="h-4 w-4" />
		{:else}
			<span aria-hidden="true">{currentFlag}</span>
		{/if}
	</Button>

	{#if open}
		<ul
			role="listbox"
			class="bg-popover text-popover-foreground absolute right-0 z-50 mt-1 min-w-[6.5rem] rounded-md border p-1 shadow-md"
		>
			{#each APP_LOCALES as loc (loc.code)}
				<li role="option" aria-selected={loc.code === current}>
					<button
						type="button"
						class="hover:bg-accent hover:text-accent-foreground flex w-full cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm
							{loc.code === current ? 'bg-accent/60' : ''}"
						onclick={() => choose(loc.code)}
					>
						<span aria-hidden="true">{loc.flag}</span>
						<span>{loc.label}</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
