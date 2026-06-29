<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { logout } from '$lib/auth';
	import { currentUser } from '$lib/stores/user';
	import { households } from '$lib/stores/households';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import * as Popover from '$lib/components/ui/popover';
	import { Sun, Moon, LayoutList, CalendarDays, Settings, Plus, LogOut, ListChecks, Users, ChevronsUpDown, Check, UserRound } from 'lucide-svelte';
	import { isDark, initTheme, toggleTheme } from '$lib/theme';
	import { isSaaS } from '$lib/env';
	import Logo from '$lib/components/Logo.svelte';

	let { onclose }: { onclose?: () => void } = $props();

	const user = $derived($currentUser);

	let switcherOpen = $state(false);
	let userMenuOpen = $state(false);

	const familyID = $derived($page.params.id);
	const currentPath = $derived($page.url.pathname);

	onMount(async () => {
		initTheme();
		const fetched = await api.get<{ id: string; name: string }[]>('/api/v1/households');
		if (fetched) households.set(fetched);
	});

	const currentFamily = $derived($households.find(f => f.id === familyID));

	function isActive(href: string) {
		return currentPath === href;
	}

	const subNav = $derived(familyID ? [
		{ label: 'Today',    href: `/households/${familyID}`,          icon: Sun,          color: 'var(--color-today)',    bg: 'var(--color-today-bg)' },
		{ label: 'Board',    href: `/households/${familyID}/board`,     icon: LayoutList,   color: 'var(--color-tasks)',    bg: 'var(--color-tasks-bg)' },
		{ label: 'Calendar', href: `/households/${familyID}/calendar`,  icon: CalendarDays, color: 'var(--color-calendar)', bg: 'var(--color-calendar-bg)' },
		{ label: 'Lists',    href: `/households/${familyID}/lists`,     icon: ListChecks,   color: 'var(--color-lists)',    bg: 'var(--color-lists-bg)' },
		{ label: 'Settings', href: `/households/${familyID}/settings`,  icon: Settings,     color: null,                   bg: null },
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
			<Logo size={24} class="text-primary shrink-0" />
			Homeboard
		</a>
	</div>

	<!-- Family switcher -->
	<div class="px-3 pt-2 pb-2 shrink-0">
		<Popover.Root bind:open={switcherOpen}>
			<Popover.Trigger
				class="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-sm text-left
					hover:bg-sidebar-accent/60 transition-colors cursor-pointer"
				aria-label="Switch household"
			>
				<Users class="w-4 h-4 shrink-0 text-muted-foreground" />
				<span class="flex-1 truncate font-medium text-sidebar-foreground">
					{currentFamily?.name ?? 'Select a family'}
				</span>
				<ChevronsUpDown class="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
			</Popover.Trigger>
			<Popover.Content class="w-56 p-1 gap-0" align="start">
				{#each $households as family (family.id)}
					<a
						href="/households/{family.id}"
						onclick={() => { switcherOpen = false; onclose?.(); }}
						class="flex items-center gap-2 px-2 py-2 rounded-xl text-sm hover:bg-accent transition-colors w-full"
					>
						<Check class="w-4 h-4 shrink-0 {family.id === familyID ? 'opacity-100' : 'opacity-0'}" />
						<span class="truncate">{family.name}</span>
					</a>
				{/each}
				{#if isSaaS || $households.length === 0}
					{#if $households.length > 0}
						<div class="my-1 h-px bg-border"></div>
					{/if}
					<a
						href="/households/new"
						onclick={() => { switcherOpen = false; onclose?.(); }}
						class="flex items-center gap-2 px-2 py-2 rounded-md text-sm hover:bg-accent transition-colors w-full text-muted-foreground"
					>
						<Plus class="w-4 h-4 shrink-0" />
						New household
					</a>
				{/if}
			</Popover.Content>
		</Popover.Root>
	</div>

	<!-- Sub-nav for current family -->
	<div class="flex-1 overflow-y-auto px-3 py-3 flex flex-col gap-0.5">
		{#if subNav.length > 0}
			{#each subNav as item (item.href)}
				{@const Icon = item.icon}
				{@const active = isActive(item.href)}
				<a
					href={item.href}
					onclick={onclose}
					aria-current={active ? 'page' : undefined}
					class="flex items-center gap-2.5 px-2 py-2 rounded-lg text-sm transition-colors
						{active
							? 'bg-sidebar-accent/40 text-sidebar-foreground font-medium'
							: 'text-sidebar-foreground hover:bg-sidebar-accent/60'}"
				>
					<span
						class="flex items-center justify-center w-7 h-7 rounded-lg shrink-0 transition-colors"
						style={item.color ? `background-color: ${active ? item.color : item.bg}; color: ${active ? 'white' : item.color};` : 'opacity: 0.6;'}
					>
						<Icon class="w-4 h-4" />
					</span>
					{item.label}
				</a>
			{/each}
		{:else}
			<p class="px-2 text-xs text-muted-foreground mt-1">Select a household to get started.</p>
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
					<button
						onclick={toggleTheme}
						class="flex items-center gap-2 px-2 py-2 rounded-xl text-sm hover:bg-accent transition-colors w-full text-left"
					>
						{#if $isDark}
							<Sun class="w-4 h-4 shrink-0 opacity-70" />
							Light mode
						{:else}
							<Moon class="w-4 h-4 shrink-0 opacity-70" />
							Dark mode
						{/if}
					</button>
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
