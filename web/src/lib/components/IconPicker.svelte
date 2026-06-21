<script lang="ts">
	import * as Popover from '$lib/components/ui/popover';
	import { Smile } from 'lucide-svelte';

	let { value = $bindable<string | undefined>(undefined) }: { value?: string } = $props();

	const EMOJIS = [
		'✅','📝','🛒','🏠','💼','🎯','⚡','🔧','📞','💊',
		'🚗','✈️','🎉','💰','📚','❤️','🌟','⚽','🎮','🍕',
		'🏃','💪','🎨','🔑','🎂','🎄','🎃','🎓','💍','🎸',
		'🍽️','🎬','🏥','🤝','🎁','🧹','🌿','🐾','📅','⏰',
	];

	let open = $state(false);
</script>

<Popover.Root bind:open>
	<Popover.Trigger>
		<button
			type="button"
			class="w-9 h-9 flex items-center justify-center rounded-md border border-input text-base hover:bg-accent transition-colors shrink-0 cursor-pointer"
			aria-label="Pick icon"
		>
			{#if value}
				{value}
			{:else}
				<Smile class="w-4 h-4 text-muted-foreground" />
			{/if}
		</button>
	</Popover.Trigger>
	<Popover.Content class="w-auto p-2" align="start">
		<div class="grid grid-cols-10 gap-0.5">
			{#each EMOJIS as emoji}
				<button
					type="button"
					class="w-7 h-7 flex items-center justify-center rounded text-sm hover:bg-accent transition-colors cursor-pointer {value === emoji ? 'bg-accent' : ''}"
					onclick={() => { value = value === emoji ? undefined : emoji; open = false; }}
				>{emoji}</button>
			{/each}
		</div>
		{#if value}
			<button
				type="button"
				class="w-full mt-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors py-0.5 cursor-pointer"
				onclick={() => { value = undefined; open = false; }}
			>Clear</button>
		{/if}
	</Popover.Content>
</Popover.Root>
