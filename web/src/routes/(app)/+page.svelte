<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { allowMultiHousehold } from '$lib/stores/config';
	import { getLastHouseholdId } from '$lib/stores/lastHousehold';
	import { Input } from '$lib/components/ui/input';
	import { Card, CardHeader, CardTitle } from '$lib/components/ui/card';
	import * as msg from '$lib/paraglide/messages.js';

	type Family = { id: string; name: string; created_at: string };

	let families = $state<Family[]>([]);
	let loading = $state(true);
	let inviteInput = $state('');

	onMount(async () => {
		let fetched: Family[] = [];
		try {
			fetched = (await api.get<Family[]>('/api/v1/households')) ?? [];
		} catch { }
		families = fetched;

		const lastId = getLastHouseholdId();
		if (lastId && fetched.some(f => f.id === lastId)) {
			await goto(`${base}/households/${lastId}`);
			return;
		}

		loading = false;
	});

	function joinWithLink() {
		const input = inviteInput.trim();
		try {
			const url = new URL(input);
			const match = url.pathname.match(/\/invite\/([^/]+)/);
			if (match) return goto(`${base}/invite/${match[1]}`);
		} catch {}
		if (input) goto(`${base}/invite/${input}`);
	}
</script>

<div class="px-4 md:px-6 pt-6 pb-8">
<div class="max-w-lg mx-auto flex flex-col gap-8">
	<div>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold">{msg.households_title()}</h2>
			{#if $allowMultiHousehold || families.length === 0}<Button href="{base}/households/new" size="sm">{msg.sidebar_new_household()}</Button>{/if}
		</div>

		{#if loading}
			<p class="text-sm text-muted-foreground">{msg.invite_loading()}</p>
		{:else if families.length === 0}
			<p class="text-sm text-muted-foreground">{msg.households_empty()}</p>
		{:else}
			<ul class="flex flex-col gap-2">
				{#each families as family (family.id)}
					<li>
						<a href="{base}/households/{family.id}">
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
		<h3 class="text-sm font-medium mb-3">{msg.households_join_heading()}</h3>
		<form onsubmit={(e) => { e.preventDefault(); joinWithLink(); }} class="flex gap-2">
			<Input
				bind:value={inviteInput}
				placeholder={msg.households_join_placeholder()}
				class="flex-1"
			/>
			<Button type="submit" variant="outline">{msg.households_join_button()}</Button>
		</form>
	</div>
</div>
</div>
