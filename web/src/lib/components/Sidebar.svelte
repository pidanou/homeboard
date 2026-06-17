<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { logout } from '$lib/auth';
	import { Sun, LayoutList, CalendarDays, Settings, Plus, LogOut, ListChecks, Users } from 'lucide-svelte';

	let { onclose }: { onclose?: () => void } = $props();

	type Family = { id: string; name: string };
	let families = $state<Family[]>([]);

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
	<div class="px-4 py-4 border-b border-sidebar-border shrink-0">
		<a
			href="/"
			class="flex items-center gap-2 font-bold text-base text-sidebar-foreground hover:opacity-80 transition-opacity"
			onclick={onclose}
		>
			<span class="text-xl">🏠</span>
			Family Board
		</a>
	</div>

	<div class="flex-1 overflow-y-auto px-3 py-3 flex flex-col gap-1">
		<!-- Families section -->
		<div class="flex items-center justify-between px-2 mb-1 mt-1">
			<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Families</span>
			<a
				href="/families/new"
				class="p-1 rounded-md hover:bg-sidebar-accent text-muted-foreground hover:text-sidebar-accent-foreground transition-colors"
				title="New family"
				onclick={onclose}
			>
				<Plus class="w-3.5 h-3.5" />
			</a>
		</div>

		{#if families.length === 0}
			<p class="px-2 text-xs text-muted-foreground">No families yet.</p>
		{/if}

		{#each families as family (family.id)}
			<a
				href="/families/{family.id}"
				onclick={onclose}
				class="flex items-center gap-2.5 px-2 py-2 rounded-lg text-sm transition-colors
					{family.id === familyID
						? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
						: 'text-sidebar-foreground hover:bg-sidebar-accent/60'}"
			>
				<Users class="w-4 h-4 shrink-0 opacity-70" />
				<span class="truncate">{family.name}</span>
			</a>
		{/each}

		<!-- Family sub-nav -->
		{#if familyID && subNav.length > 0}
			<div class="mt-2 pt-3 border-t border-sidebar-border flex flex-col gap-0.5">
				<div class="px-2 mb-1">
					<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider truncate block">
						{currentFamily?.name ?? '…'}
					</span>
				</div>
				{#each subNav as item (item.href)}
					{@const Icon = item.icon}
					<a
						href={item.href}
						onclick={onclose}
						class="flex items-center gap-2.5 px-2 py-2 rounded-lg text-sm transition-colors
							{isActive(item.href)
								? 'bg-sidebar-primary text-sidebar-primary-foreground font-medium'
								: 'text-sidebar-foreground hover:bg-sidebar-accent/60'}"
					>
						<Icon class="w-4 h-4 shrink-0" />
						{item.label}
					</a>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Sign out -->
	<div class="px-3 pb-4 pt-2 border-t border-sidebar-border shrink-0">
		<button
			onclick={logout}
			class="flex items-center gap-2.5 px-2 py-2 rounded-lg text-sm w-full text-left
				text-sidebar-foreground hover:bg-sidebar-accent/60 transition-colors"
		>
			<LogOut class="w-4 h-4 shrink-0 opacity-70" />
			Sign out
		</button>
	</div>
</div>
