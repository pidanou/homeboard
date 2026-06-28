<script lang="ts">
    import { page } from "$app/stores";
    import { onMount } from "svelte";
    import { api } from "$lib/api/client";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { X, Pencil, Clock } from "lucide-svelte";
    import UserAvatar from "$lib/components/UserAvatar.svelte";
    import { currentUser } from "$lib/stores/user";
    import { households, updateHouseholdName } from "$lib/stores/households";
    import { subscribePush, unsubscribePush, isPushSubscribed } from "$lib/push";

    type Invite = { token: string; expires_at: string };
    type Member = {
        user_id: string;
        name: string;
        email: string;
        avatar_url?: string | null;
        role: string;
        joined_at: string;
        virtual?: boolean;
    };
    type CategoryColor =
        | "red"
        | "orange"
        | "yellow"
        | "green"
        | "teal"
        | "blue"
        | "purple"
        | "pink"
        | "gray";
    type AppCategory = { id: string; name: string; color: CategoryColor };

    const CATEGORY_COLORS: CategoryColor[] = [
        "red",
        "orange",
        "yellow",
        "green",
        "teal",
        "blue",
        "purple",
        "pink",
        "gray",
    ];
    const CATEGORY_DOT: Record<CategoryColor, string> = {
        red: "bg-rose-500",
        orange: "bg-orange-400",
        yellow: "bg-amber-400",
        green: "bg-emerald-600",
        teal: "bg-teal-600",
        blue: "bg-indigo-500",
        purple: "bg-violet-500",
        pink: "bg-pink-400",
        gray: "bg-stone-400",
    };

    const familyID = $derived($page.params.id);
    const householdName = $derived(
        $households.find((h) => h.id === familyID)?.name ?? "",
    );

    let invite = $state<Invite | null>(null);
    let members = $state<Member[]>([]);
    let categories = $state<AppCategory[]>([]);
    let copied = $state<string | null>(null);
    let pushSubscribed = $state(false);
    let pushSupported = $state(false);

    // name editing
    let editingName = $state(false);
    let editNameValue = $state("");

    // category state
    let newCategoryName = $state("");
    let newCategoryColor = $state<CategoryColor>("blue");
    let addingVirtual = $state(false);
    let newVirtualName = $state("");
    let editingCatID = $state<string | null>(null);
    let editingCatName = $state("");
    let editingCatColor = $state<CategoryColor>("blue");

    const myRole = $derived(
        members.find((m) => m.user_id === $currentUser?.id)?.role ?? "member",
    );
    const isAdmin = $derived(myRole === "admin");
    const realCount = $derived(members.filter((m) => !m.virtual).length);
    const virtualCount = $derived(members.filter((m) => m.virtual).length);

    onMount(async () => {
        const [membersResult, invitesResult, categoriesResult] =
            await Promise.allSettled([
                api.get<Member[]>(`/api/v1/households/${familyID}/members`),
                api.get<Invite[]>(`/api/v1/households/${familyID}/invites`),
                api.get<AppCategory[]>(
                    `/api/v1/households/${familyID}/categories`,
                ),
            ]);
        if (membersResult.status === "fulfilled")
            members = membersResult.value ?? [];
        if (invitesResult.status === "fulfilled")
            invite = (invitesResult.value ?? [])[0] ?? null;
        if (categoriesResult.status === "fulfilled")
            categories = categoriesResult.value ?? [];
        pushSupported = 'serviceWorker' in navigator && 'PushManager' in window;
        if (pushSupported) pushSubscribed = await isPushSubscribed();
    });

    async function saveName() {
        const trimmed = editNameValue.trim();
        if (!trimmed || trimmed === householdName || !familyID) {
            editingName = false;
            return;
        }
        try {
            await api.patch(`/api/v1/households/${familyID}`, {
                name: trimmed,
            });
            updateHouseholdName(familyID, trimmed);
        } catch {}
        editingName = false;
    }

    function startEditName() {
        editNameValue = householdName;
        editingName = true;
    }

    async function createCategory() {
        if (!newCategoryName.trim()) return;
        try {
            const cat = await api.post<AppCategory>(
                `/api/v1/households/${familyID}/categories`,
                {
                    name: newCategoryName.trim(),
                    color: newCategoryColor,
                },
            );
            categories = [...categories, cat];
            newCategoryName = "";
        } catch {}
    }

    async function deleteCategory(categoryID: string) {
        try {
            await api.delete(
                `/api/v1/households/${familyID}/categories/${categoryID}`,
            );
            categories = categories.filter((c) => c.id !== categoryID);
        } catch {}
    }

    async function createVirtualMember() {
        if (!newVirtualName.trim()) return;
        try {
            const vm = await api.post<{ id: string; name: string }>(
                `/api/v1/households/${familyID}/members/virtual`,
                { name: newVirtualName.trim() },
            );
            members = [
                ...members,
                {
                    user_id: vm.id,
                    name: vm.name,
                    email: "",
                    role: "",
                    joined_at: "",
                    virtual: true,
                },
            ];
            newVirtualName = "";
            addingVirtual = false;
        } catch {}
    }

    async function deleteVirtualMember(id: string) {
        try {
            await api.delete(
                `/api/v1/households/${familyID}/members/virtual/${id}`,
            );
            members = members.filter((m) => m.user_id !== id);
        } catch {}
    }

    async function updateRole(userID: string, role: "admin" | "member") {
        try {
            await api.put(
                `/api/v1/households/${familyID}/members/${userID}/role`,
                { role },
            );
            members = members.map((m) =>
                m.user_id === userID ? { ...m, role } : m,
            );
        } catch {}
    }

    async function kickMember(userID: string) {
        try {
            await api.delete(
                `/api/v1/households/${familyID}/members/${userID}`,
            );
            members = members.filter((m) => m.user_id !== userID);
        } catch {}
    }

    function startEditCat(cat: AppCategory) {
        editingCatID = cat.id;
        editingCatName = cat.name;
        editingCatColor = cat.color;
    }

    async function saveEditCat(cat: AppCategory) {
        if (!editingCatName.trim()) return;
        try {
            await api.put(
                `/api/v1/households/${familyID}/categories/${cat.id}`,
                {
                    name: editingCatName.trim(),
                    color: editingCatColor,
                },
            );
            categories = categories.map((c) =>
                c.id === cat.id
                    ? {
                          ...c,
                          name: editingCatName.trim(),
                          color: editingCatColor,
                      }
                    : c,
            );
            editingCatID = null;
        } catch {}
    }

    async function generateInvite() {
        try {
            invite = await api.post<Invite>(
                `/api/v1/households/${familyID}/invites`,
                {},
            );
        } catch {}
    }

    async function revokeInvite() {
        try {
            if (!invite) return;
            await api.delete(
                `/api/v1/households/${familyID}/invites/${invite.token}`,
            );
            invite = null;
        } catch {}
    }

    function copyLink(token: string) {
        navigator.clipboard.writeText(`${location.origin}/invite/${token}`);
        copied = token;
        setTimeout(() => (copied = null), 2000);
    }

    function initials(name: string) {
        return name
            .split(" ")
            .map((w) => w[0])
            .join("")
            .slice(0, 2)
            .toUpperCase();
    }
</script>

<div class="px-4 md:px-6 pt-4 md:pt-6 pb-12">
    <div class="max-w-2xl mx-auto flex flex-col gap-0 divide-y divide-border">
        <div class="pb-6">
            <h1 class="text-2xl font-bold">Settings</h1>
        </div>

        <!-- General -->
        <section class="py-6 flex flex-col gap-4">
            <h2
                class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
            >
                General
            </h2>

            <div
                class="rounded-xl border border-border bg-card overflow-hidden"
            >
                <div class="flex items-center gap-3 px-4 py-3.5">
                    <span class="text-sm text-muted-foreground w-20 shrink-0"
                        >Name</span
                    >
                    {#if editingName}
                        <div class="flex flex-1 items-center gap-2">
                            <Input
                                bind:value={editNameValue}
                                class="flex-1 h-8 text-sm"
                                autofocus
                                onkeydown={(e) => {
                                    if (e.key === "Enter") {
                                        e.preventDefault();
                                        saveName();
                                    }
                                    if (e.key === "Escape") {
                                        editingName = false;
                                    }
                                }}
                            />
                            <Button
                                size="sm"
                                onclick={saveName}
                                disabled={!editNameValue.trim()}
                                class="h-8">Save</Button
                            >
                            <Button
                                size="sm"
                                variant="ghost"
                                onclick={() => (editingName = false)}
                                class="h-8">Cancel</Button
                            >
                        </div>
                    {:else}
                        <span class="flex-1 text-sm font-medium"
                            >{householdName}</span
                        >
                        {#if isAdmin}
                            <Button
                                size="sm"
                                variant="ghost"
                                onclick={startEditName}
                                class="h-8 gap-1.5 text-muted-foreground"
                            >
                                <Pencil class="w-3.5 h-3.5" />
                            </Button>
                        {/if}
                    {/if}
                </div>
            </div>
        </section>

        <!-- Members -->
        <section class="py-6 flex flex-col gap-4">
            <div class="flex items-center justify-between gap-3">
                <div>
                    <h2
                        class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
                    >
                        Members
                    </h2>
                    <p class="text-xs text-muted-foreground mt-0.5">
                        {realCount} member{realCount !== 1 ? "s" : ""}{virtualCount > 0
                            ? ` · ${virtualCount} profile${virtualCount !== 1 ? "s" : ""}`
                            : ""}
                    </p>
                </div>
                {#if isAdmin}
                    <Button
                        size="sm"
                        variant="outline"
                        onclick={() => (addingVirtual = !addingVirtual)}
                    >
                        Add profile
                    </Button>
                {/if}
            </div>

            {#if addingVirtual}
                <div class="flex gap-2">
                    <Input
                        bind:value={newVirtualName}
                        placeholder="Name (e.g. Lucas)…"
                        class="flex-1"
                        onkeydown={(e) => {
                            if (e.key === "Enter") {
                                e.preventDefault();
                                createVirtualMember();
                            }
                            if (e.key === "Escape") addingVirtual = false;
                        }}
                    />
                    <Button
                        size="sm"
                        onclick={createVirtualMember}
                        disabled={!newVirtualName.trim()}>Add</Button
                    >
                    <Button
                        size="sm"
                        variant="ghost"
                        onclick={() => (addingVirtual = false)}>Cancel</Button
                    >
                </div>
            {/if}

            {#if members.length === 0}
                <p class="text-sm text-muted-foreground">No members yet.</p>
            {:else}
                <div
                    class="rounded-xl border border-border bg-card overflow-hidden divide-y divide-border"
                >
                    {#each members as member (member.user_id)}
                        <div class="flex items-center gap-3 px-4 py-3">
                            <UserAvatar
                                name={member.name}
                                avatarUrl={member.virtual
                                    ? null
                                    : member.avatar_url}
                                userId={member.user_id}
                                size={32}
                            />
                            <div class="flex-1 min-w-0">
                                <p class="text-sm font-medium truncate">
                                    {member.name}
                                </p>
                                <p
                                    class="text-xs text-muted-foreground truncate"
                                >
                                    {#if member.virtual}Profile{:else}{member.email}{/if}
                                </p>
                            </div>
                            {#if member.virtual}
                                <div class="flex items-center gap-2 shrink-0">
                                    <span class="text-xs px-2 py-0.5 rounded-full font-medium bg-muted text-muted-foreground">
                                        Profile
                                    </span>
                                    {#if isAdmin}
                                        <button
                                            onclick={() =>
                                                deleteVirtualMember(member.user_id)}
                                            class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                                            aria-label="Remove"
                                        >
                                            <X class="w-4 h-4" />
                                        </button>
                                    {/if}
                                </div>
                            {:else}
                                <div class="flex items-center gap-2 shrink-0">
                                    <span
                                        class="text-xs px-2 py-0.5 rounded-full font-medium
									{member.role === 'admin'
                                            ? 'bg-primary/10 text-primary'
                                            : 'bg-muted text-muted-foreground'}"
                                    >
                                        {member.role === "admin"
                                            ? "Admin"
                                            : "Member"}
                                    </span>
                                    {#if isAdmin && member.user_id !== $currentUser?.id}
                                        <Button
                                            size="sm"
                                            variant="outline"
                                            onclick={() =>
                                                updateRole(
                                                    member.user_id,
                                                    member.role === "admin"
                                                        ? "member"
                                                        : "admin",
                                                )}
                                            class="h-7 px-2 text-xs"
                                        >
                                            {member.role === "admin"
                                                ? "Demote"
                                                : "Make admin"}
                                        </Button>
                                        <button
                                            onclick={() =>
                                                kickMember(member.user_id)}
                                            class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                                            aria-label="Remove member"
                                        >
                                            <X class="w-4 h-4" />
                                        </button>
                                    {/if}
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            {/if}
        </section>

        <!-- Categories -->
        <section class="py-6 flex flex-col gap-4">
            <h2
                class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
            >
                Categories
            </h2>

            {#if categories.length > 0}
                <div
                    class="rounded-xl border border-border bg-card overflow-hidden divide-y divide-border"
                >
                    {#each categories as cat (cat.id)}
                        <div class="px-4 py-3">
                            {#if editingCatID === cat.id}
                                <div class="flex flex-col gap-2.5">
                                    <Input
                                        bind:value={editingCatName}
                                        class="h-8 text-sm"
                                        onkeydown={(e) => {
                                            if (e.key === "Enter") {
                                                e.preventDefault();
                                                saveEditCat(cat);
                                            }
                                            if (e.key === "Escape") {
                                                editingCatID = null;
                                            }
                                        }}
                                    />
                                    <div class="flex gap-1.5">
                                        {#each CATEGORY_COLORS as c}
                                            <button
                                                type="button"
                                                title={c}
                                                onclick={() =>
                                                    (editingCatColor = c)}
                                                class="w-5 h-5 rounded-full {CATEGORY_DOT[
                                                    c
                                                ]} transition-all
												{editingCatColor === c
                                                    ? 'ring-2 ring-offset-1 ring-foreground'
                                                    : 'opacity-50 hover:opacity-90'}"
                                            ></button>
                                        {/each}
                                    </div>
                                    <div class="flex gap-2">
                                        <Button
                                            size="sm"
                                            onclick={() => saveEditCat(cat)}
                                            disabled={!editingCatName.trim()}
                                            >Save</Button
                                        >
                                        <Button
                                            size="sm"
                                            variant="ghost"
                                            onclick={() =>
                                                (editingCatID = null)}
                                            >Cancel</Button
                                        >
                                    </div>
                                </div>
                            {:else}
                                <div
                                    class="flex items-center justify-between gap-2"
                                >
                                    <span
                                        class="flex items-center gap-2.5 text-sm"
                                    >
                                        <span
                                            class="w-2.5 h-2.5 rounded-full {CATEGORY_DOT[
                                                cat.color
                                            ]} shrink-0"
                                        ></span>
                                        {cat.name}
                                    </span>
                                    {#if isAdmin}
                                        <div class="flex items-center gap-1">
                                            <button
                                                onclick={() =>
                                                    startEditCat(cat)}
                                                class="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                                                aria-label="Edit"
                                            >
                                                <Pencil class="w-3.5 h-3.5" />
                                            </button>
                                            <button
                                                onclick={() =>
                                                    deleteCategory(cat.id)}
                                                class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
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
                <div
                    class="rounded-xl border border-border bg-card p-4 flex flex-col gap-3"
                >
                    <p
                        class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
                    >
                        New category
                    </p>
                    <Input
                        bind:value={newCategoryName}
                        placeholder="Category name…"
                        onkeydown={(e) => {
                            if (e.key === "Enter") {
                                e.preventDefault();
                                createCategory();
                            }
                        }}
                    />
                    <div class="flex gap-2">
                        {#each CATEGORY_COLORS as c}
                            <button
                                type="button"
                                onclick={() => (newCategoryColor = c)}
                                class="w-6 h-6 rounded-full {CATEGORY_DOT[
                                    c
                                ]} transition-all
								{newCategoryColor === c
                                    ? 'ring-2 ring-offset-2 ring-foreground'
                                    : 'opacity-50 hover:opacity-90'}"
                                title={c}
                            ></button>
                        {/each}
                    </div>
                    <Button
                        onclick={createCategory}
                        disabled={!newCategoryName.trim()}
                        size="sm"
                    >
                        Add category
                    </Button>
                </div>
            {/if}
        </section>

        <!-- Invite link -->
        {#if isAdmin}
            <section class="py-6 flex flex-col gap-4">
                <div class="flex items-center justify-between gap-3">
                    <h2
                        class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
                    >
                        Invite link
                    </h2>
                    <Button
                        size="sm"
                        variant="outline"
                        onclick={generateInvite}
                    >
                        {invite ? "Regenerate" : "Generate link"}
                    </Button>
                </div>

                {#if invite}
                    {@const daysLeft = Math.ceil(
                        (new Date(invite.expires_at).getTime() - Date.now()) /
                            86400000,
                    )}
                    <div
                        class="rounded-xl border border-border bg-card overflow-hidden"
                    >
                        <div
                            class="flex items-center gap-2 px-4 py-3 text-xs font-mono text-muted-foreground border-b border-border"
                        >
                            <span class="flex-1 truncate"
                                >{location.origin}/invite/{invite.token}</span
                            >
                            <span
                                class="inline-flex items-center gap-1 shrink-0 font-sans font-medium
						{daysLeft <= 1
                                    ? 'text-destructive'
                                    : daysLeft <= 3
                                      ? 'text-amber-600 dark:text-amber-400'
                                      : 'text-muted-foreground'}"
                            >
                                <Clock class="w-3 h-3" />
                                {daysLeft <= 0
                                    ? "Expires today"
                                    : `${daysLeft}d left`}
                            </span>
                        </div>
                        <div class="flex gap-2 px-4 py-3">
                            <Button
                                variant="outline"
                                size="sm"
                                class="flex-1"
                                onclick={() => copyLink(invite!.token)}
                            >
                                {copied === invite.token
                                    ? "Copied!"
                                    : "Copy link"}
                            </Button>
                            <Button
                                variant="destructive"
                                size="sm"
                                onclick={revokeInvite}>Revoke</Button
                            >
                        </div>
                    </div>
                {:else}
                    <p class="text-sm text-muted-foreground">
                        No active link. Generate one to invite someone.
                    </p>
                {/if}
            </section>
        {/if}
    {#if pushSupported}
        <section class="space-y-3">
            <h2 class="text-lg font-semibold">Notifications</h2>
            <p class="text-sm text-muted-foreground">
                Get notified when new events or tasks are added to this family.
            </p>
            <Button
                variant={pushSubscribed ? "outline" : "default"}
                onclick={async () => {
                    if (pushSubscribed) {
                        await unsubscribePush(familyID);
                        pushSubscribed = false;
                    } else {
                        await subscribePush(familyID);
                        pushSubscribed = await isPushSubscribed();
                    }
                }}
            >
                {pushSubscribed ? "Disable notifications" : "Enable notifications"}
            </Button>
        </section>
    {/if}
    </div>
</div>
