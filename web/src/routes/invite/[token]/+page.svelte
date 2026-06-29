<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { isLoggedIn, setToken } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';

	type Invite = { token: string; family_id: string; family_name: string; expires_at: string };
	type VirtualMember = { id: string; name: string };
	type AcceptResult = { family_id: string; unlinked_virtual_members: VirtualMember[] | null };
	type RegisterResult = AcceptResult & { token: string };

	const token = $derived($page.params.token);

	let invite = $state<Invite | null>(null);
	let failed = $state(false);
	let loading = $state(false);
	let unlinked = $state<VirtualMember[]>([]);
	let familyID = $state('');

	// Registration form (shown when not logged in)
	let name = $state('');
	let email = $state('');
	let password = $state('');
	let regError = $state('');

	onMount(async () => {
		try {
			invite = await api.get<Invite>(`/api/v1/invites/${token}`);
		} catch {
			failed = true;
		}
	});

	async function accept() {
		loading = true;
		try {
			const result = await api.post<AcceptResult>(`/api/v1/invites/${token}/accept`, {});
			familyID = result.family_id;
			const members = result.unlinked_virtual_members ?? [];
			if (members.length > 0) {
				unlinked = members;
			} else {
				goto('/');
			}
		} catch { } finally {
			loading = false;
		}
	}

	async function registerAndAccept() {
		regError = '';
		loading = true;
		try {
			const result = await api.post<RegisterResult>(`/api/v1/invites/${token}/register`, { name, email, password });
			setToken(result.token);
			familyID = result.family_id;
			const members = result.unlinked_virtual_members ?? [];
			if (members.length > 0) {
				unlinked = members;
			} else {
				goto('/');
			}
		} catch {
			regError = 'Registration failed. The email may already be taken.';
		} finally {
			loading = false;
		}
	}

	async function link(virtualID: string) {
		try {
			await api.post(`/api/v1/households/${familyID}/members/virtual/${virtualID}/link`, {});
		} catch {
			// non-fatal — still redirect
		}
		goto('/');
	}
</script>

<div class="min-h-screen flex items-center justify-center px-4">
	<div class="max-w-sm w-full text-center flex flex-col gap-4">
		<h1 class="text-2xl font-bold">Homeboard</h1>

		{#if failed}
			<p class="text-destructive text-sm">Invite not found or expired.</p>

		{:else if unlinked.length > 0}
			<!-- Linking prompt -->
			<p class="text-muted-foreground text-sm">Are you one of these people already in the household?</p>
			<div class="flex flex-col gap-2 text-left">
				{#each unlinked as vm (vm.id)}
					<button
						onclick={() => link(vm.id)}
						class="flex items-center gap-3 px-4 py-3 rounded-lg border border-border bg-card hover:bg-muted transition-colors text-sm font-medium cursor-pointer"
					>
						<span class="w-8 h-8 rounded-full bg-primary/15 text-primary flex items-center justify-center text-xs font-semibold shrink-0">
							{vm.name.slice(0, 2).toUpperCase()}
						</span>
						{vm.name}
					</button>
				{/each}
				<button
					onclick={() => goto('/')}
					class="px-4 py-2 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
				>
					No, I'm someone new
				</button>
			</div>

		{:else if invite && isLoggedIn()}
			<p class="text-muted-foreground">You've been invited to join <span class="font-semibold text-foreground">{invite.family_name}</span>.</p>
			<Button onclick={accept} disabled={loading} class="w-full">
				{loading ? 'Accepting…' : 'Accept invite'}
			</Button>

		{:else if invite}
			<p class="text-muted-foreground">You've been invited to join <span class="font-semibold text-foreground">{invite.family_name}</span>.</p>
			<p class="text-sm text-muted-foreground">Create an account to continue.</p>

			<form onsubmit={(e) => { e.preventDefault(); registerAndAccept(); }} class="flex flex-col gap-3 text-left">
				<div class="flex flex-col gap-1">
					<Label for="name">Name</Label>
					<Input id="name" bind:value={name} placeholder="Your name" required />
				</div>
				<div class="flex flex-col gap-1">
					<Label for="email">Email</Label>
					<Input id="email" type="email" bind:value={email} placeholder="you@example.com" required />
				</div>
				<div class="flex flex-col gap-1">
					<Label for="password">Password</Label>
					<Input id="password" type="password" bind:value={password} placeholder="••••••••" required minlength={8} />
				</div>
				{#if regError}
					<p class="text-destructive text-sm">{regError}</p>
				{/if}
				<Button type="submit" disabled={loading} class="w-full">
					{loading ? 'Creating account…' : 'Create account & join'}
				</Button>
			</form>

			<p class="text-sm text-muted-foreground">
				Already have an account?
				<a href="/login?redirect=/invite/{token}" class="underline">Sign in</a>
			</p>

		{:else}
			<p class="text-sm text-muted-foreground">Loading…</p>
		{/if}
	</div>
</div>
