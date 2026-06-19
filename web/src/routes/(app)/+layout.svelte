<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isLoggedIn } from '$lib/auth';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import { currentUser, loadCurrentUser } from '$lib/stores/user';
	import { Menu, Sun, LayoutList, CalendarDays, ListChecks, Settings } from 'lucide-svelte';

	let { children } = $props();
	let ready = $state(false);
	let mobileMenuOpen = $state(false);

	const familyID = $derived($page.params.id);
	const currentPath = $derived($page.url.pathname);

	// Close mobile menu on navigation
	$effect(() => {
		currentPath;
		mobileMenuOpen = false;
	});

	const user = $derived($currentUser);

	onMount(() => {
		if (!isLoggedIn()) {
			goto('/login');
		} else {
			ready = true;
			loadCurrentUser();
		}
	});

	const mobileTabNav = $derived(familyID ? [
		{ label: 'Today',    href: `/families/${familyID}`,           icon: Sun },
		{ label: 'Board',    href: `/families/${familyID}/board`,     icon: LayoutList },
		{ label: 'Calendar', href: `/families/${familyID}/calendar`,  icon: CalendarDays },
		{ label: 'Lists',    href: `/families/${familyID}/lists`,     icon: ListChecks },
		{ label: 'Settings', href: `/families/${familyID}/settings`,  icon: Settings },
	] : []);

	const currentSection = $derived(() => {
		if (!familyID) return currentPath === '/profile' ? 'Profile' : 'Family Board';
		if (currentPath === `/families/${familyID}`) return 'Today';
		if (currentPath.endsWith('/board')) return 'Board';
		if (currentPath.endsWith('/calendar')) return 'Calendar';
		if (currentPath.endsWith('/lists')) return 'Lists';
		if (currentPath.endsWith('/settings')) return 'Settings';
		return 'Family Board';
	});
</script>

{#if ready}
	<div class="h-screen flex bg-background overflow-hidden">
		<!-- Desktop sidebar (always visible md+) -->
		<aside aria-label="Main navigation" class="hidden md:flex w-56 shrink-0 flex-col border-r border-sidebar-border bg-sidebar fixed top-0 left-0 bottom-0 z-30">
			<Sidebar />
		</aside>

		<!-- Mobile sidebar overlay -->
		{#if mobileMenuOpen}
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div class="fixed inset-0 z-40 md:hidden" onclick={() => mobileMenuOpen = false}>
				<div class="absolute inset-0 bg-black/40"></div>
				<aside aria-label="Main navigation" class="absolute left-0 top-0 bottom-0 w-64 bg-sidebar flex flex-col z-50" onclick={(e) => e.stopPropagation()}>
					<Sidebar onclose={() => mobileMenuOpen = false} />
				</aside>
			</div>
		{/if}

		<!-- Main area -->
		<div class="flex-1 flex flex-col min-w-0 md:ml-56">
			<!-- Mobile top bar -->
			<header class="md:hidden sticky top-0 z-20 border-b border-border bg-background/95 backdrop-blur-sm px-4 h-14 flex items-center justify-between shrink-0">
				<span class="font-semibold text-base">{currentSection()}</span>
				<div class="flex items-center gap-1">
					{#if user}
						<a href="/profile" class="p-1 rounded-full hover:opacity-80 transition-opacity" aria-label="My profile">
							<UserAvatar name={user.name} avatarUrl={user.avatar_url} userId={user.id} size={32} />
						</a>
					{/if}
					<button
						onclick={() => mobileMenuOpen = !mobileMenuOpen}
						class="p-2 rounded-lg hover:bg-muted transition-colors"
						aria-label="Open menu"
					>
						<Menu class="w-5 h-5" />
					</button>
				</div>
			</header>

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
