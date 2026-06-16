<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Card, CardHeader, CardTitle } from '$lib/components/ui/card';

	type Family = { id: string; name: string; created_at: string };

	let families = $state<Family[]>([]);
	let loading = $state(true);
	let error = $state('');
	let inviteInput = $state('');

	onMount(async () => {
		try {
			families = (await api.get<Family[]>('/api/v1/families')) ?? [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load families';
		} finally {
			loading = false;
		}
	});

	function joinWithLink() {
		const input = inviteInput.trim();
		try {
			const url = new URL(input);
			const match = url.pathname.match(/\/invite\/([^/]+)/);
			if (match) return goto(`/invite/${match[1]}`);
		} catch {}
		if (input) goto(`/invite/${input}`);
	}
</script>

<div class="max-w-lg mx-auto flex flex-col gap-8">
	<div>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold">Your families</h2>
			<Button href="/families/new" size="sm">New family</Button>
		</div>

		{#if loading}
			<p class="text-sm text-muted-foreground">Loading…</p>
		{:else if error}
			<p class="text-sm text-destructive">{error}</p>
		{:else if families.length === 0}
			<p class="text-sm text-muted-foreground">No families yet. Create one or join with an invite link.</p>
		{:else}
			<ul class="flex flex-col gap-2">
				{#each families as family (family.id)}
					<li>
						<a href="/families/{family.id}">
							<Card class="hover:bg-muted/50 transition-colors cursor-pointer">
								<CardHeader>
									<CardTitle class="text-base">{family.name}</CardTitle>
								</CardHeader>
							</Card>
						</a>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	<div class="border-t pt-6">
		<h3 class="text-sm font-medium mb-3">Join with an invite link</h3>
		<form onsubmit={(e) => { e.preventDefault(); joinWithLink(); }} class="flex gap-2">
			<Input
				bind:value={inviteInput}
				placeholder="Paste invite link or token…"
				class="flex-1"
			/>
			<Button type="submit" variant="outline">Join</Button>
		</form>
	</div>
</div>
