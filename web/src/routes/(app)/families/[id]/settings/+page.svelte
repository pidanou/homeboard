<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';

	type Invite = { token: string; expires_at: string };

	const familyID = $derived($page.params.id);

	let invites = $state<Invite[]>([]);
	let error = $state('');
	let copied = $state<string | null>(null);

	onMount(async () => {
		try {
			invites = (await api.get<Invite[]>(`/api/v1/families/${familyID}/invites`)) ?? [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load invites';
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

{#if error}
	<p class="text-sm text-destructive">{error}</p>
{/if}

<div class="flex flex-col gap-4">
	<div class="flex items-center justify-between">
		<h3 class="text-sm font-medium">Invite links</h3>
		<Button size="sm" variant="outline" onclick={createInvite}>Generate new</Button>
	</div>

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
					<Button variant="destructive" size="sm" onclick={() => revokeInvite(invite.token)}>Revoke</Button>
				</div>
				<p class="text-xs text-muted-foreground">Expires {new Date(invite.expires_at).toLocaleDateString()}</p>
			</div>
		{/each}
	{/if}
</div>
