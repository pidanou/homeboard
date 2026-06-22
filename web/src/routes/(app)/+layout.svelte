<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isLoggedIn } from '$lib/auth';
	import { api } from '$lib/api/client';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import { currentUser, loadCurrentUser } from '$lib/stores/user';
	import { Sun, LayoutList, CalendarDays, ListChecks, Settings, ChevronDown } from 'lucide-svelte';

	let { children } = $props();
	let ready = $state(false);
	let offline = $state(false);
	let householdName = $state<string | null>(null);

	const familyID = $derived($page.params.id);
	const currentPath = $derived($page.url.pathname);

	const user = $derived($currentUser);

	$effect(() => {
		if (familyID) {
			api.get<{ id: string; name: string }>(`/api/v1/households/${familyID}`)
				.then(h => { householdName = h?.name ?? null; })
				.catch(() => { householdName = null; });
		} else {
			householdName = null;
		}
	});

	onMount(() => {
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
		<aside aria-label="Main navigation" class="hidden md:flex w-56 shrink-0 flex-col border-r border-sidebar-border bg-sidebar fixed top-0 left-0 bottom-0 z-30">
			<Sidebar />
		</aside>

		<!-- Main area -->
		<div class="flex-1 flex flex-col min-w-0 md:ml-56">
			<!-- Mobile top bar -->
			<header class="md:hidden sticky top-0 z-20 border-b border-border bg-background/95 backdrop-blur-sm px-4 safe-area-top flex flex-col shrink-0">
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

			<main class="flex-1 overflow-auto">
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
