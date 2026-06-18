<script lang="ts">
    import type { Member, AppCategory, Filter } from "$lib/types";
    import { dotClass } from "$lib/categories";
    import { X, Tag } from "lucide-svelte";

    let {
        filter = $bindable<Filter>("all"),
        filterMembers = $bindable(new Set<string>()),
        filterCategory = $bindable<string | null>(null),
        members,
        categories,
        doneCnt,
        familyID,
    }: {
        filter?: Filter;
        filterMembers?: Set<string>;
        filterCategory?: string | null;
        members: Member[];
        categories: AppCategory[];
        doneCnt: number;
        familyID: string;
    } = $props();

    const someFilterActive = $derived(
        filter !== "all" || filterMembers.size > 0 || filterCategory !== null,
    );

    function chipCls(active: boolean) {
        if (active) return "ring-1 ring-foreground opacity-100";
        return someFilterActive ? "opacity-30" : "opacity-70 hover:opacity-100";
    }

    function initials(name: string) {
        return name
            .split(" ")
            .map((w) => w[0])
            .join("")
            .slice(0, 2)
            .toUpperCase();
    }

    function toggleMember(uid: string) {
        const next = new Set(filterMembers);
        next.has(uid) ? next.delete(uid) : next.add(uid);
        filterMembers = next;
    }

    const FILTERS: { id: Filter; label: string }[] = $derived([
        { id: "all", label: "Active" },
        { id: "tasks", label: "Tasks" },
        { id: "events", label: "Events" },
        { id: "done", label: `Done${doneCnt ? ` (${doneCnt})` : ""}` },
    ]);
</script>

<div class="flex items-center gap-2 flex-wrap">
    <!-- Type pills -->
    {#each FILTERS as f}
        <button
            onclick={() => (filter = f.id)}
            class="px-2 py-0.5 rounded-full text-xs font-medium transition-all cursor-pointer {chipCls(
                filter === f.id,
            )}">{f.label}</button
        >
    {/each}

    {#if categories.length > 0}
        <span class="text-border text-xs hidden sm:block">|</span>
        {#each categories as cat (cat.id)}
            <button
                onclick={() => {
                    filterCategory = filterCategory === cat.id ? null : cat.id;
                }}
                class="hidden sm:flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs transition-all cursor-pointer {chipCls(
                    filterCategory === cat.id,
                )}"
            >
                <span
                    class="w-2 h-2 rounded-full {dotClass(cat.color)} shrink-0"
                ></span>
                {cat.name}
            </button>
        {/each}
    {/if}

    {#if members.length > 0}
        <span class="text-border text-xs hidden sm:block">|</span>
        {#each members as m (m.user_id)}
            <button
                onclick={() => toggleMember(m.user_id)}
                title={m.name}
                class="hidden sm:flex w-6 h-6 rounded-full text-[10px] font-semibold items-center justify-center transition-all cursor-pointer shrink-0
					{filterMembers.has(m.user_id)
                    ? 'bg-primary text-primary-foreground ring-1 ring-foreground'
                    : someFilterActive
                      ? 'bg-muted text-muted-foreground opacity-30'
                      : 'bg-muted text-muted-foreground opacity-70 hover:opacity-100'}"
                >{initials(m.name)}</button
            >
        {/each}
    {/if}

    {#if someFilterActive}
        <button
            onclick={() => {
                filter = "all";
                filterMembers = new Set();
                filterCategory = null;
            }}
            class="flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            ><X class="w-3 h-3" />Clear</button
        >
    {/if}
</div>
