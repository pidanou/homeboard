<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { X, UserX } from 'lucide-svelte';

	type Invite = { token: string; expires_at: string };
	type Member = { user_id: string; name: string; email: string; role: string; joined_at: string; virtual?: boolean };
	type LabelColor = 'red' | 'orange' | 'yellow' | 'green' | 'teal' | 'blue' | 'purple' | 'pink' | 'gray';
	type AppLabel = { id: string; name: string; color: LabelColor };

	const LABEL_COLORS: LabelColor[] = ['red', 'orange', 'yellow', 'green', 'teal', 'blue', 'purple', 'pink', 'gray'];
	const LABEL_DOT: Record<LabelColor, string> = {
		red: 'bg-red-500', orange: 'bg-orange-500', yellow: 'bg-yellow-500',
		green: 'bg-green-500', teal: 'bg-teal-500', blue: 'bg-blue-500',
		purple: 'bg-purple-500', pink: 'bg-pink-500', gray: 'bg-gray-400',
	};

	const familyID = $derived($page.params.id);

	let invites = $state<Invite[]>([]);
	let members = $state<Member[]>([]);
	let labels = $state<AppLabel[]>([]);
	let error = $state('');
	let copied = $state<string | null>(null);
	let newLabelName = $state('');
	let newLabelColor = $state<LabelColor>('blue');
	let addingVirtual = $state(false);
	let newVirtualName = $state('');

	onMount(async () => {
		const [membersResult, invitesResult, labelsResult] = await Promise.allSettled([
			api.get<Member[]>(`/api/v1/families/${familyID}/members`),
			api.get<Invite[]>(`/api/v1/families/${familyID}/invites`),
			api.get<AppLabel[]>(`/api/v1/families/${familyID}/labels`),
		]);
		if (membersResult.status === 'fulfilled') members = membersResult.value ?? [];
		if (invitesResult.status === 'fulfilled') invites = invitesResult.value ?? [];
		if (labelsResult.status === 'fulfilled') labels = labelsResult.value ?? [];
		if (membersResult.status === 'rejected' || invitesResult.status === 'rejected') {
			error = 'Failed to load settings';
		}
	});

	async function createLabel() {
		if (!newLabelName.trim()) return;
		try {
			const label = await api.post<AppLabel>(`/api/v1/families/${familyID}/labels`, {
				name: newLabelName.trim(),
				color: newLabelColor,
			});
			labels = [...labels, label];
			newLabelName = '';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create label';
		}
	}

	async function deleteLabel(labelID: string) {
		try {
			await api.delete(`/api/v1/families/${familyID}/labels/${labelID}`);
			labels = labels.filter((l) => l.id !== labelID);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete label';
		}
	}

	async function createVirtualMember() {
		if (!newVirtualName.trim()) return;
		try {
			const vm = await api.post<{ id: string; name: string }>(`/api/v1/families/${familyID}/members/virtual`, { name: newVirtualName.trim() });
			members = [...members, { user_id: vm.id, name: vm.name, email: '', role: '', virtual: true }];
			newVirtualName = '';
			addingVirtual = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to add member';
		}
	}

	async function deleteVirtualMember(id: string) {
		try {
			await api.delete(`/api/v1/families/${familyID}/members/virtual/${id}`);
			members = members.filter((m) => m.user_id !== id);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to remove member';
		}
	}

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

	function initials(name: string) {
		return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
	}
</script>

{#if error}
	<p class="text-sm text-destructive mb-4">{error}</p>
{/if}

<div class="flex flex-col gap-8">

	<!-- Members -->
	<div class="flex flex-col gap-3">
		<div class="flex items-center justify-between">
			<h3 class="text-sm font-semibold">Members ({members.length})</h3>
			<Button size="sm" variant="outline" onclick={() => (addingVirtual = !addingVirtual)}>
				+ Without account
			</Button>
		</div>

		{#if addingVirtual}
			<div class="flex gap-2">
				<Input
					bind:value={newVirtualName}
					placeholder="Name (e.g. Lucas)…"
					class="flex-1"
					onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); createVirtualMember(); } if (e.key === 'Escape') addingVirtual = false; }}
				/>
				<Button size="sm" onclick={createVirtualMember} disabled={!newVirtualName.trim()}>Add</Button>
				<Button size="sm" variant="ghost" onclick={() => (addingVirtual = false)}>Cancel</Button>
			</div>
		{/if}

		{#if members.length === 0}
			<p class="text-sm text-muted-foreground">No members yet.</p>
		{:else}
			<div class="flex flex-col gap-2">
				{#each members as member (member.user_id)}
					<div class="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-3">
						<div class="w-8 h-8 rounded-full {member.virtual ? 'bg-muted text-muted-foreground' : 'bg-primary/15 text-primary'} flex items-center justify-center text-xs font-semibold shrink-0">
							{#if member.virtual}
								<UserX class="w-4 h-4" />
							{:else}
								{initials(member.name)}
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium truncate">{member.name}</p>
							<p class="text-xs text-muted-foreground truncate">
								{#if member.virtual}
									No account
								{:else}
									{member.email}
								{/if}
							</p>
						</div>
						{#if member.virtual}
							<button
								onclick={() => deleteVirtualMember(member.user_id)}
								class="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
								aria-label="Remove"
							>
								<X class="w-3.5 h-3.5" />
							</button>
						{:else}
							<span class="text-xs text-muted-foreground capitalize">{member.role}</span>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Labels -->
	<div class="flex flex-col gap-3">
		<h3 class="text-sm font-semibold">Labels</h3>
		{#if labels.length > 0}
			<div class="flex flex-col gap-1.5">
				{#each labels as lbl (lbl.id)}
					<div class="flex items-center justify-between gap-2 rounded-lg border border-border bg-card px-4 py-2.5">
						<span class="flex items-center gap-2 text-sm">
							<span class="w-3 h-3 rounded-full {LABEL_DOT[lbl.color]} shrink-0"></span>
							{lbl.name}
						</span>
						<button
							onclick={() => deleteLabel(lbl.id)}
							class="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
						>
							<X class="w-3.5 h-3.5" />
						</button>
					</div>
				{/each}
			</div>
		{:else}
			<p class="text-sm text-muted-foreground">No labels yet.</p>
		{/if}

		<div class="flex flex-col gap-2 rounded-lg border border-border bg-card p-4">
			<p class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">New label</p>
			<Input
				bind:value={newLabelName}
				placeholder="Label name…"
				onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); createLabel(); } }}
			/>
			<div class="flex flex-wrap gap-2">
				{#each LABEL_COLORS as c}
					<button
						type="button"
						onclick={() => (newLabelColor = c)}
						class="w-6 h-6 rounded-full {LABEL_DOT[c]} transition-all
							{newLabelColor === c ? 'ring-2 ring-offset-2 ring-foreground' : 'opacity-70 hover:opacity-100'}"
						title={c}
					></button>
				{/each}
			</div>
			<Button onclick={createLabel} disabled={!newLabelName.trim()} size="sm" class="w-full">
				Add label
			</Button>
		</div>
	</div>

	<!-- Invite links -->
	<div class="flex flex-col gap-3">
		<div class="flex items-center justify-between">
			<h3 class="text-sm font-semibold">Invite links</h3>
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

</div>
