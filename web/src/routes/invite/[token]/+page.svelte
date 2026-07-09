<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
	import { api } from '$lib/api/client';
	import { isLoggedIn, setToken } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as m from '$lib/paraglide/messages.js';

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
				goto(`${base}/`);
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
				goto(`${base}/`);
			}
		} catch {
			regError = m.invite_register_failed();
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
		goto(`${base}/`);
	}
</script>

<div class="min-h-screen flex items-center justify-center px-4">
	<div class="max-w-sm w-full text-center flex flex-col gap-4">
		<h1 class="text-2xl font-bold">Homeboard</h1>

		{#if failed}
			<p class="text-destructive text-sm">{m.invite_not_found()}</p>

		{:else if unlinked.length > 0}
			<!-- Linking prompt -->
			<p class="text-muted-foreground text-sm">{m.invite_link_question()}</p>
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
					onclick={() => goto(`${base}/`)}
					class="px-4 py-2 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
				>
					{m.invite_someone_new()}
				</button>
			</div>

		{:else if invite && isLoggedIn()}
			<p class="text-muted-foreground">{m.invite_join_prompt({ familyName: invite.family_name })}</p>
			<Button onclick={accept} disabled={loading} class="w-full">
				{loading ? m.invite_accepting() : m.invite_accept()}
			</Button>

		{:else if invite}
			<p class="text-muted-foreground">{m.invite_join_prompt({ familyName: invite.family_name })}</p>
			<p class="text-sm text-muted-foreground">{m.invite_create_account_prompt()}</p>

			<form onsubmit={(e) => { e.preventDefault(); registerAndAccept(); }} class="flex flex-col gap-3 text-left">
				<div class="flex flex-col gap-1">
					<Label for="name">{m.auth_name_label()}</Label>
					<Input id="name" bind:value={name} placeholder={m.invite_name_placeholder()} required />
				</div>
				<div class="flex flex-col gap-1">
					<Label for="email">{m.auth_email_label()}</Label>
					<Input id="email" type="email" bind:value={email} placeholder={m.invite_email_placeholder()} required />
				</div>
				<div class="flex flex-col gap-1">
					<Label for="password">{m.auth_password_label()}</Label>
					<Input id="password" type="password" bind:value={password} placeholder="••••••••" required minlength={8} />
				</div>
				{#if regError}
					<p class="text-destructive text-sm">{regError}</p>
				{/if}
				<Button type="submit" disabled={loading} class="w-full">
					{loading ? m.invite_creating_account() : m.invite_create_join()}
				</Button>
			</form>

			<p class="text-sm text-muted-foreground">
				{m.invite_have_account()}
				<a href="{base}/login?redirect={base}/invite/{token}" class="underline">{m.invite_sign_in()}</a>
			</p>

		{:else}
			<p class="text-sm text-muted-foreground">{m.invite_loading()}</p>
		{/if}
	</div>
</div>
