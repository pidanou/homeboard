<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { isLoggedIn } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';

	type Invite = { token: string; family_id: string; family_name: string; expires_at: string };

	const token = $derived($page.params.token);

	let invite = $state<Invite | null>(null);
	let error = $state('');
	let loading = $state(false);

	onMount(async () => {
		try {
			invite = await api.get<Invite>(`/api/v1/invites/${token}`);
		} catch {
			error = 'Invite not found or expired.';
		}
	});

	async function accept() {
		if (!isLoggedIn()) {
			goto(`/login?redirect=/invite/${token}`);
			return;
		}
		loading = true;
		try {
			await api.post(`/api/v1/invites/${token}/accept`, {});
			goto('/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to accept invite';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center px-4">
	<div class="max-w-sm w-full text-center flex flex-col gap-4">
		<h1 class="text-2xl font-bold">Family Board</h1>
		{#if error}
			<p class="text-destructive text-sm">{error}</p>
		{:else if invite}
			<p class="text-muted-foreground">You've been invited to join <span class="font-semibold text-foreground">{invite.family_name}</span>.</p>
			<Button onclick={accept} disabled={loading} class="w-full">
				{loading ? 'Accepting…' : 'Accept invite'}
			</Button>
		{:else}
			<p class="text-sm text-muted-foreground">Loading…</p>
		{/if}
	</div>
</div>
