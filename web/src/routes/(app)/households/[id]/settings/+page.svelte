<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { X, Pencil, Clock } from 'lucide-svelte';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import { currentUser } from '$lib/stores/user';

	type Invite = { token: string; expires_at: string };
	type Member = { user_id: string; name: string; email: string; avatar_url?: string | null; role: string; joined_at: string; virtual?: boolean };
	type CategoryColor = 'red' | 'orange' | 'yellow' | 'green' | 'teal' | 'blue' | 'purple' | 'pink' | 'gray';
	type AppCategory = { id: string; name: string; color: CategoryColor };

	const CATEGORY_COLORS: CategoryColor[] = ['red', 'orange', 'yellow', 'green', 'teal', 'blue', 'purple', 'pink', 'gray'];
	const CATEGORY_DOT: Record<CategoryColor, string> = {
		red: 'bg-rose-500', orange: 'bg-orange-400', yellow: 'bg-amber-400',
		green: 'bg-emerald-600', teal: 'bg-teal-600', blue: 'bg-indigo-500',
		purple: 'bg-violet-500', pink: 'bg-pink-400', gray: 'bg-stone-400',
	};

	const familyID = $derived($page.params.id);

	let invite = $state<Invite | null>(null);
	let members = $state<Member[]>([]);

	const myRole = $derived(members.find((m) => m.user_id === $currentUser?.id)?.role ?? 'member');
	const isAdmin = $derived(myRole === 'admin');
	const realCount = $derived(members.filter((m) => !m.virtual).length);
	const virtualCount = $derived(members.filter((m) => m.virtual).length);
	let categories = $state<AppCategory[]>([]);
	let copied = $state<string | null>(null);
	let newCategoryName = $state('');
	let newCategoryColor = $state<CategoryColor>('blue');
	let addingVirtual = $state(false);
	let newVirtualName = $state('');
	let editingCatID = $state<string | null>(null);
	let editingCatName = $state('');
	let editingCatColor = $state<CategoryColor>('blue');

	onMount(async () => {
		const [membersResult, invitesResult, categoriesResult] = await Promise.allSettled([
			api.get<Member[]>(`/api/v1/households/${familyID}/members`),
			api.get<Invite[]>(`/api/v1/households/${familyID}/invites`),
			api.get<AppCategory[]>(`/api/v1/households/${familyID}/categories`),
		]);
		if (membersResult.status === 'fulfilled') members = membersResult.value ?? [];
		if (invitesResult.status === 'fulfilled') invite = (invitesResult.value ?? [])[0] ?? null;
		if (categoriesResult.status === 'fulfilled') categories = categoriesResult.value ?? [];
	});

	async function createCategory() {
		if (!newCategoryName.trim()) return;
		try {
			const cat = await api.post<AppCategory>(`/api/v1/households/${familyID}/categories`, {
				name: newCategoryName.trim(),
				color: newCategoryColor,
			});
			categories = [...categories, cat];
			newCategoryName = '';
		} catch { }
	}

	async function deleteCategory(categoryID: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/categories/${categoryID}`);
			categories = categories.filter((c) => c.id !== categoryID);
		} catch { }
	}

	async function createVirtualMember() {
		if (!newVirtualName.trim()) return;
		try {
			const vm = await api.post<{ id: string; name: string }>(`/api/v1/households/${familyID}/members/virtual`, { name: newVirtualName.trim() });
			members = [...members, { user_id: vm.id, name: vm.name, email: '', role: '', joined_at: '', virtual: true }];
			newVirtualName = '';
			addingVirtual = false;
		} catch { }
	}

	async function deleteVirtualMember(id: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/members/virtual/${id}`);
			members = members.filter((m) => m.user_id !== id);
		} catch { }
	}

	async function updateRole(userID: string, role: 'admin' | 'member') {
		try {
			await api.put(`/api/v1/households/${familyID}/members/${userID}/role`, { role });
			members = members.map((m) => (m.user_id === userID ? { ...m, role } : m));
		} catch {}
	}

	async function kickMember(userID: string) {
		try {
			await api.delete(`/api/v1/households/${familyID}/members/${userID}`);
			members = members.filter((m) => m.user_id !== userID);
		} catch { }
	}

	function startEditCat(cat: AppCategory) {
		editingCatID = cat.id;
		editingCatName = cat.name;
		editingCatColor = cat.color;
	}

	async function saveEditCat(cat: AppCategory) {
		if (!editingCatName.trim()) return;
		try {
			await api.put(`/api/v1/households/${familyID}/categories/${cat.id}`, {
				name: editingCatName.trim(),
				color: editingCatColor,
			});
			categories = categories.map((c) =>
				c.id === cat.id ? { ...c, name: editingCatName.trim(), color: editingCatColor } : c
			);
			editingCatID = null;
		} catch { }
	}

	async function generateInvite() {
		try {
			invite = await api.post<Invite>(`/api/v1/households/${familyID}/invites`, {});
		} catch { }
	}

	async function revokeInvite() {
		try {
			if (!invite) return;
			await api.delete(`/api/v1/households/${familyID}/invites/${invite.token}`);
			invite = null;
		} catch { }
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

<div class="flex flex-col gap-8 pt-4 md:pt-6 px-4 md:px-6 pb-8">

	<!-- Members -->
	<div class="flex flex-col gap-3">
		<div class="flex items-start justify-between gap-3">
			<div>
				<h3 class="text-sm font-semibold">Members</h3>
				<p class="text-xs text-muted-foreground">
					{realCount} with account{virtualCount > 0 ? ` · ${virtualCount} without` : ''}
				</p>
			</div>
			{#if isAdmin}
				<Button size="sm" variant="outline" class="shrink-0" onclick={() => (addingVirtual = !addingVirtual)}>
					Add without account
				</Button>
			{/if}
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
						<UserAvatar name={member.name} avatarUrl={member.virtual ? null : member.avatar_url} userId={member.user_id} size={32} />
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
							{#if isAdmin}
								<button
									onclick={() => deleteVirtualMember(member.user_id)}
									class="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
									aria-label="Remove"
								>
									<X class="w-3.5 h-3.5" />
								</button>
							{/if}
						{:else}
							<div class="flex flex-wrap items-center justify-end gap-1 shrink-0">
								<span class="text-xs px-1.5 py-0.5 rounded-full font-medium
									{member.role === 'admin' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}">
									{member.role === 'admin' ? 'Admin' : 'Member'}
								</span>
								{#if isAdmin && member.user_id !== $currentUser?.id}
									<Button
										size="sm"
										variant="outline"
										onclick={() => updateRole(member.user_id, member.role === 'admin' ? 'member' : 'admin')}
										class="h-6 px-2 text-xs"
									>
										{member.role === 'admin' ? 'Make member' : 'Make admin'}
									</Button>
									<button
										onclick={() => kickMember(member.user_id)}
										class="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
										aria-label="Kick member"
									>
										<X class="w-3.5 h-3.5" />
									</button>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Categories -->
	<div class="flex flex-col gap-3">
		<h3 class="text-sm font-semibold">Categories</h3>
		{#if categories.length > 0}
			<div class="flex flex-col gap-1.5">
				{#each categories as cat (cat.id)}
					<div class="flex flex-col gap-2 rounded-lg border border-border bg-card px-4 py-2.5">
						{#if editingCatID === cat.id}
							<Input
								bind:value={editingCatName}
								class="h-7 text-sm"
								onkeydown={(e) => {
									if (e.key === 'Enter') { e.preventDefault(); saveEditCat(cat); }
									if (e.key === 'Escape') { editingCatID = null; }
								}}
							/>
							<div class="flex flex-wrap gap-1.5">
								{#each CATEGORY_COLORS as c}
									<button
										type="button"
										onclick={() => (editingCatColor = c)}
										class="w-5 h-5 rounded-full {CATEGORY_DOT[c]} transition-all
											{editingCatColor === c ? 'ring-2 ring-offset-1 ring-foreground' : 'opacity-60 hover:opacity-100'}"
									></button>
								{/each}
							</div>
							<div class="flex gap-2">
								<Button size="sm" onclick={() => saveEditCat(cat)} disabled={!editingCatName.trim()} class="flex-1">Save</Button>
								<Button size="sm" variant="ghost" onclick={() => (editingCatID = null)}>Cancel</Button>
							</div>
						{:else}
							<div class="group flex items-center justify-between gap-2">
								<span class="flex items-center gap-2 text-sm">
									<span class="w-3 h-3 rounded-full {CATEGORY_DOT[cat.color]} shrink-0"></span>
									{cat.name}
								</span>
								{#if isAdmin}
									<div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
										<button
											onclick={() => startEditCat(cat)}
											class="p-1 rounded text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
											aria-label="Edit"
										>
											<Pencil class="w-3.5 h-3.5" />
										</button>
										<button
											onclick={() => deleteCategory(cat.id)}
											class="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
											aria-label="Delete"
										>
											<X class="w-3.5 h-3.5" />
										</button>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{:else}
			<p class="text-sm text-muted-foreground">No categories yet.</p>
		{/if}

		{#if isAdmin}
		<div class="flex flex-col gap-2 rounded-lg border border-border bg-card p-4">
			<p class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">New category</p>
			<Input
				bind:value={newCategoryName}
				placeholder="Category name…"
				onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); createCategory(); } }}
			/>
			<div class="flex flex-wrap gap-2">
				{#each CATEGORY_COLORS as c}
					<button
						type="button"
						onclick={() => (newCategoryColor = c)}
						class="w-6 h-6 rounded-full {CATEGORY_DOT[c]} transition-all
							{newCategoryColor === c ? 'ring-2 ring-offset-2 ring-foreground' : 'opacity-70 hover:opacity-100'}"
						title={c}
					></button>
				{/each}
			</div>
			<Button onclick={createCategory} disabled={!newCategoryName.trim()} size="sm" class="w-full">
				Add category
			</Button>
		</div>
		{/if}
	</div>

	<!-- Invite link — admin only -->
	{#if isAdmin}
	<div class="flex flex-col gap-3">
		<div class="flex items-center justify-between">
			<h3 class="text-sm font-semibold">Invite link</h3>
			<Button size="sm" variant="outline" onclick={generateInvite}>
				{invite ? 'Regenerate' : 'Generate link'}
			</Button>
		</div>

		{#if invite}
			{@const daysLeft = Math.ceil((new Date(invite.expires_at).getTime() - Date.now()) / 86400000)}
			<div class="flex flex-col gap-1">
				<div class="flex gap-2">
					<Input readonly value="{location.origin}/invite/{invite.token}" class="flex-1 text-xs" />
					<Button variant="outline" size="sm" onclick={() => copyLink(invite!.token)}>
						{copied === invite.token ? 'Copied!' : 'Copy'}
					</Button>
					<Button variant="destructive" size="sm" onclick={revokeInvite}>Revoke</Button>
				</div>
				<span class="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full font-medium
					{daysLeft <= 1 ? 'bg-destructive/10 text-destructive' : daysLeft <= 3 ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400' : 'bg-muted text-muted-foreground'}">
					<Clock class="w-3 h-3" />
					{daysLeft <= 0 ? 'Expires today' : `${daysLeft}d left`}
				</span>
			</div>
		{:else}
			<p class="text-sm text-muted-foreground">No active link. Generate one to invite someone.</p>
		{/if}
	</div>
	{/if}

</div>
