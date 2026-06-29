<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isLoggedIn } from '$lib/auth';
	import { api, getBaseUrl } from '$lib/api/client';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import { currentUser, loadCurrentUser } from '$lib/stores/user';
	import { households } from '$lib/stores/households';
	import { Sun, LayoutList, CalendarDays, ListChecks, Settings, ChevronDown, ChevronLeft, ChevronRight } from 'lucide-svelte';

	let { children } = $props();
	let ready = $state(false);
	let offline = $state(false);
	let household = $state<{ id: string; name: string; wallpaper_url?: string | null; wallpaper_original_url?: string | null } | null>(null);
	let wallpaperBlobUrl = $state<string | null>(null);
	let sidebarCollapsed = $state(false);

	const familyID = $derived($page.params.id);
	const currentPath = $derived($page.url.pathname);
	const householdName = $derived(household?.name ?? null);
	const wallpaperUrl = $derived($households.find(h => h.id === familyID)?.wallpaper_url ?? null);

	const user = $derived($currentUser);

	$effect(() => {
		if (familyID) {
			api.get<{ id: string; name: string; wallpaper_url?: string | null; wallpaper_original_url?: string | null }>(`/api/v1/households/${familyID}`)
				.then(h => { household = h ?? null; })
				.catch(() => { household = null; });
		} else {
			household = null;
		}
	});

	$effect(() => {
		const url = wallpaperUrl;
		if (!url) { wallpaperBlobUrl = null; return; }
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		const fullUrl = url.startsWith('/') ? `${getBaseUrl()}${url}` : url;
		let localBlob: string | null = null;
		let active = true;
		fetch(fullUrl, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
			.then(r => r.ok ? r.blob() : null)
			.then(blob => {
				if (!active) return;
				localBlob = blob ? URL.createObjectURL(blob) : null;
				wallpaperBlobUrl = localBlob;
			})
			.catch(() => { if (active) wallpaperBlobUrl = null; });
		return () => {
			active = false;
			if (localBlob) URL.revokeObjectURL(localBlob);
		};
	});

	onMount(() => {
		sidebarCollapsed = localStorage.getItem('sidebar-collapsed') === 'true';

		if (!isLoggedIn()) {
			goto('/login');
		} else {
			ready = true;
			loadCurrentUser();
		}

		offline = !navigator.onLine;
		const goOffline = () => (offline = true);
		const goOnline = () => window.location.reload();
		window.addEventListener('offline', goOffline);
		window.addEventListener('online', goOnline);
		return () => {
			window.removeEventListener('offline', goOffline);
			window.removeEventListener('online', goOnline);
		};
	});

	function toggleSidebar() {
		sidebarCollapsed = !sidebarCollapsed;
		localStorage.setItem('sidebar-collapsed', String(sidebarCollapsed));
	}

	const mobileTabNav = $derived(familyID ? [
		{ label: 'Today',    href: `/households/${familyID}`,           icon: Sun },
		{ label: 'Board',    href: `/households/${familyID}/board`,     icon: LayoutList },
		{ label: 'Calendar', href: `/households/${familyID}/calendar`,  icon: CalendarDays },
		{ label: 'Lists',    href: `/households/${familyID}/lists`,     icon: ListChecks },
		{ label: 'Settings', href: `/households/${familyID}/settings`,  icon: Settings },
	] : []);

	const currentSection = $derived(() => {
		if (!familyID) return currentPath === '/profile' ? 'Profile' : 'Homeboard';
		if (currentPath === `/households/${familyID}`) return 'Today';
		if (currentPath.endsWith('/board')) return 'Board';
		if (currentPath.endsWith('/calendar')) return 'Calendar';
		if (currentPath.endsWith('/lists')) return 'Lists';
		if (currentPath.endsWith('/settings')) return 'Settings';
		return 'Homeboard';
	});
</script>

{#if ready}
	<div class="h-dvh flex bg-background overflow-hidden">
		<!-- Desktop sidebar (always visible md+) -->
		<aside
			aria-label="Main navigation"
			class="hidden md:flex shrink-0 flex-col border-r border-sidebar-border bg-sidebar fixed top-0 left-0 bottom-0 z-30 transition-[width] duration-200
				{sidebarCollapsed ? 'w-14' : 'w-56'}"
		>
			<Sidebar collapsed={sidebarCollapsed} ontoggle={toggleSidebar} />
			<button
				onclick={toggleSidebar}
				aria-label={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
				class="absolute right-0 top-6 translate-x-1/2 z-40 flex items-center justify-center w-6 h-6 rounded-full border border-sidebar-border bg-sidebar text-muted-foreground hover:text-foreground shadow-sm transition-colors"
			>
				{#if sidebarCollapsed}
					<ChevronRight class="w-3 h-3" />
				{:else}
					<ChevronLeft class="w-3 h-3" />
				{/if}
			</button>
		</aside>

		<!-- Main area -->
		<div class="flex-1 flex flex-col min-w-0 relative transition-[margin] duration-200 {sidebarCollapsed ? 'md:ml-14' : 'md:ml-56'}">
			{#if wallpaperBlobUrl}
			<div
				class="absolute inset-0 bg-cover bg-center pointer-events-none z-0 opacity-35 mix-blend-multiply dark:mix-blend-screen dark:opacity-20"
				style="background-image: url('{wallpaperBlobUrl}');"
			></div>
			{/if}
			<!-- Mobile top bar -->
			<header class="md:hidden sticky top-0 z-20 border-b border-border bg-background/70 backdrop-blur-sm px-4 safe-area-top flex flex-col shrink-0">
				<div class="h-14 flex items-center justify-between w-full">
					{#if householdName}
						<a href="/" class="flex items-center gap-1 font-semibold text-base truncate max-w-[65%] hover:opacity-70 transition-opacity">
						<span class="truncate">{householdName}</span>
						<ChevronDown class="w-4 h-4 shrink-0 text-muted-foreground" />
					</a>
					{:else}
						<span class="font-semibold text-base">{currentSection()}</span>
					{/if}
					{#if user}
						<a href="/profile" class="p-1 rounded-full hover:opacity-80 transition-opacity shrink-0" aria-label="My profile">
							<UserAvatar name={user.name} avatarUrl={user.avatar_url} userId={user.id} size={32} />
						</a>
					{/if}
				</div>
			</header>

			{#if offline}
				<div class="bg-yellow-500/90 text-yellow-950 text-xs font-medium text-center py-1.5 px-4 shrink-0">
					No internet connection
				</div>
			{/if}

			<main class="flex-1 overflow-hidden">
				{@render children()}
			</main>

			<!-- Mobile bottom tab bar (only when in a family) -->
			{#if mobileTabNav.length > 0}
				<nav aria-label="Section navigation" class="md:hidden border-t border-border bg-background shrink-0 flex safe-area-bottom">
					{#each mobileTabNav as item (item.href)}
						{@const Icon = item.icon}
						<a
							href={item.href}
							aria-current={currentPath === item.href ? 'page' : undefined}
							class="flex-1 flex flex-col items-center justify-center gap-1 py-3 text-xs font-medium transition-colors min-h-[56px]
								{currentPath === item.href
									? 'text-primary'
									: 'text-muted-foreground hover:text-foreground'}"
						>
							<Icon class="w-5 h-5" />
							{item.label}
						</a>
					{/each}
				</nav>
			{/if}
		</div>
	</div>
{/if}
