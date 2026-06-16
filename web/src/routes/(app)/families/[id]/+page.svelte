<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';

	type Family = { id: string; name: string };
	type Invite = { token: string; expires_at: string };

	const familyID = $derived($page.params.id);

	let family = $state<Family | null>(null);
	let invites = $state<Invite[]>([]);
	let error = $state('');
	let copied = $state<string | null>(null);

	onMount(async () => {
		try {
			[family, invites] = await Promise.all([
				api.get<Family>(`/api/v1/families/${familyID}`),
				api.get<Invite[]>(`/api/v1/families/${familyID}/invites`).then(r => r ?? [])
			]);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load family';
		}
	});

	async function createInvite() {
		try {
			const inv = await api.post<Invite>(`/api/v1/families/${familyID}/invites`, {});
			invites = [inv, ...invites];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create invite';
		}
	}

	async function revokeInvite(token: string) {
		try {
			await api.delete(`/api/v1/families/${familyID}/invites/${token}`);
			invites = invites.filter(i => i.token !== token);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to revoke invite';
		}
	}

	function copyLink(token: string) {
		navigator.clipboard.writeText(`${location.origin}/invite/${token}`);
		copied = token;
		setTimeout(() => (copied = null), 2000);
	}
</script>

<div class="max-w-lg mx-auto flex flex-col gap-6">
	{#if error}
		<p class="text-sm text-destructive">{error}</p>
	{:else if family}
		<h2 class="text-xl font-semibold">{family.name}</h2>

		<Card>
			<CardHeader class="flex flex-row items-center justify-between">
				<CardTitle class="text-base">Invite links</CardTitle>
				<Button size="sm" variant="outline" onclick={createInvite}>Generate new</Button>
			</CardHeader>
			<CardContent class="flex flex-col gap-3">
				{#if invites.length === 0}
					<p class="text-sm text-muted-foreground">No active invites. Generate one to share.</p>
				{:else}
					{#each invites as invite (invite.token)}
						<div class="flex flex-col gap-1">
							<div class="flex gap-2">
								<Input readonly value="{location.origin}/invite/{invite.token}" class="flex-1 text-xs" />
								<Button variant="outline" size="sm" onclick={() => copyLink(invite.token)}>
									{copied === invite.token ? 'Copied!' : 'Copy'}
								</Button>
								<Button variant="destructive" size="sm" onclick={() => revokeInvite(invite.token)}>
									Revoke
								</Button>
							</div>
							<p class="text-xs text-muted-foreground">
								Expires {new Date(invite.expires_at).toLocaleDateString()}
							</p>
						</div>
					{/each}
				{/if}
			</CardContent>
		</Card>
	{/if}
</div>
