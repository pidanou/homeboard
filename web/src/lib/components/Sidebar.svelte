<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { logout } from '$lib/auth';
	import { currentUser } from '$lib/stores/user';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import * as Popover from '$lib/components/ui/popover';
	import { Sun, LayoutList, CalendarDays, Settings, Plus, LogOut, ListChecks, Users, ChevronsUpDown, Check, UserRound } from 'lucide-svelte';

	let { onclose }: { onclose?: () => void } = $props();

	const user = $derived($currentUser);

	type Family = { id: string; name: string };
	let families = $state<Family[]>([]);
	let switcherOpen = $state(false);
	let userMenuOpen = $state(false);

	const familyID = $derived($page.params.id);
	const currentPath = $derived($page.url.pathname);

	onMount(async () => {
		families = (await api.get<Family[]>('/api/v1/families')) ?? [];
	});

	const currentFamily = $derived(families.find(f => f.id === familyID));

	function isActive(href: string) {
		return currentPath === href;
	}

	const subNav = $derived(familyID ? [
		{ label: 'Today',    href: `/families/${familyID}`,          icon: Sun },
		{ label: 'Board',    href: `/families/${familyID}/board`,     icon: LayoutList },
		{ label: 'Calendar', href: `/families/${familyID}/calendar`,  icon: CalendarDays },
		{ label: 'Lists',    href: `/families/${familyID}/lists`,     icon: ListChecks },
		{ label: 'Settings', href: `/families/${familyID}/settings`,  icon: Settings },
	] : []);
</script>

<div class="flex flex-col h-full select-none">
	<!-- Logo -->
	<div class="px-4 py-4 shrink-0">
		<a
			href="/"
			class="flex items-center gap-2 font-bold text-base text-sidebar-foreground hover:opacity-80 transition-opacity"
			onclick={onclose}
		>
			<span class="text-xl">🏠</span>
			Family Board
		</a>
	</div>

	<!-- Family switcher -->
	<div class="px-3 pt-2 pb-2 shrink-0">
		<Popover.Root bind:open={switcherOpen}>
			<Popover.Trigger
				class="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-sm text-left
					hover:bg-sidebar-accent/60 transition-colors cursor-pointer"
				aria-label="Switch family"
			>
				<Users class="w-4 h-4 shrink-0 text-muted-foreground" />
				<span class="flex-1 truncate font-medium text-sidebar-foreground">
					{currentFamily?.name ?? 'Select a family'}
				</span>
				<ChevronsUpDown class="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
			</Popover.Trigger>
			<Popover.Content class="w-56 p-1 gap-0" align="start">
				{#each families as family (family.id)}
					<a
						href="/families/{family.id}"
						onclick={() => { switcherOpen = false; onclose?.(); }}
						class="flex items-center gap-2 px-2 py-2 rounded-xl text-sm hover:bg-accent transition-colors w-full"
					>
						<Check class="w-4 h-4 shrink-0 {family.id === familyID ? 'opacity-100' : 'opacity-0'}" />
						<span class="truncate">{family.name}</span>
					</a>
				{/each}
				{#if families.length > 0}
					<div class="my-1 h-px bg-border"></div>
				{/if}
				<a
					href="/families/new"
					onclick={() => { switcherOpen = false; onclose?.(); }}
					class="flex items-center gap-2 px-2 py-2 rounded-md text-sm hover:bg-accent transition-colors w-full text-muted-foreground"
				>
					<Plus class="w-4 h-4 shrink-0" />
					New family
				</a>
			</Popover.Content>
		</Popover.Root>
	</div>

	<!-- Sub-nav for current family -->
	<div class="flex-1 overflow-y-auto px-3 py-3 flex flex-col gap-0.5">
		{#if subNav.length > 0}
			{#each subNav as item (item.href)}
				{@const Icon = item.icon}
				<a
					href={item.href}
					onclick={onclose}
					aria-current={isActive(item.href) ? 'page' : undefined}
					class="flex items-center gap-2.5 px-2 py-2 rounded-lg text-sm transition-colors
						{isActive(item.href)
							? 'bg-sidebar-primary text-sidebar-primary-foreground font-medium'
							: 'text-sidebar-foreground hover:bg-sidebar-accent/60'}"
				>
					<Icon class="w-4 h-4 shrink-0" />
					{item.label}
				</a>
			{/each}
		{:else}
			<p class="px-2 text-xs text-muted-foreground mt-1">Select a family to get started.</p>
		{/if}
	</div>

	<!-- User menu -->
	<div class="px-3 pb-4 pt-2 shrink-0">
		{#if user}
			<Popover.Root bind:open={userMenuOpen}>
				<Popover.Trigger
					class="flex items-center gap-2.5 px-2 py-2 rounded-lg text-sm w-full text-left
						hover:bg-sidebar-accent/60 transition-colors cursor-pointer"
					aria-label="User menu"
				>
					<UserAvatar name={user.name} avatarUrl={user.avatar_url} userId={user.id} size={24} />
					<span class="truncate flex-1 text-sidebar-foreground">{user.name}</span>
					<ChevronsUpDown class="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
				</Popover.Trigger>
				<Popover.Content class="w-48 p-1 gap-0" align="start" side="top">
					<a
						href="/profile"
						onclick={() => { userMenuOpen = false; onclose?.(); }}
						class="flex items-center gap-2 px-2 py-2 rounded-xl text-sm hover:bg-accent transition-colors w-full"
					>
						<UserRound class="w-4 h-4 shrink-0 opacity-70" />
						Profile
					</a>
					<div class="my-1 h-px bg-border"></div>
					<button
						onclick={logout}
						class="flex items-center gap-2 px-2 py-2 rounded-xl text-sm hover:bg-accent transition-colors w-full text-left text-destructive"
					>
						<LogOut class="w-4 h-4 shrink-0" />
						Sign out
					</button>
				</Popover.Content>
			</Popover.Root>
		{/if}
	</div>
</div>
