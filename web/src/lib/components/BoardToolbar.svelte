<script lang="ts">
    import type { Member, AppCategory } from "$lib/types";
    import { dotStyle } from "$lib/categories";
    import { X, Tag } from "lucide-svelte";
    import UserAvatar from "$lib/components/UserAvatar.svelte";
    import * as msg from "$lib/paraglide/messages.js";

    type ItemType = "task" | "event" | "birthday";

    let {
        filterTypes = $bindable(new Set<ItemType>()),
        showDone = $bindable(false),
        filterMembers = $bindable(new Set<string>()),
        filterCategory = $bindable(new Set<string>()),
        members,
        categories,
        doneCnt,
        familyID,
    }: {
        filterTypes?: Set<ItemType>;
        showDone?: boolean;
        filterMembers?: Set<string>;
        filterCategory?: Set<string>;
        members: Member[];
        categories: AppCategory[];
        doneCnt: number;
        familyID: string;
    } = $props();

    const someFilterActive = $derived(
        filterTypes.size > 0 || filterMembers.size > 0 || filterCategory.size > 0,
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

    function toggleType(t: ItemType) {
        const next = new Set(filterTypes);
        next.has(t) ? next.delete(t) : next.add(t);
        filterTypes = next;
    }

    const TYPES: { id: ItemType; label: string }[] = $derived([
        { id: "task", label: msg.board_filter_tasks() },
        { id: "event", label: msg.board_filter_events() },
        { id: "birthday", label: `🎂 ${msg.board_filter_birthdays()}` },
    ]);
</script>

<div class="flex items-center gap-2 flex-wrap">
    <!-- Type pills -->
    {#each TYPES as t}
        <button
            onclick={() => toggleType(t.id)}
            class="px-2 py-0.5 rounded-full text-xs font-medium transition-all cursor-pointer {chipCls(
                filterTypes.has(t.id),
            )}">{t.label}</button
        >
    {/each}

    <span class="text-border text-xs hidden sm:block">|</span>
    <button
        onclick={() => (showDone = !showDone)}
        class="px-2 py-0.5 rounded-full text-xs font-medium transition-all cursor-pointer {showDone
            ? 'ring-1 ring-foreground opacity-100'
            : someFilterActive
              ? 'opacity-30'
              : 'opacity-70 hover:opacity-100'}"
        >{msg.board_filter_done()}{doneCnt ? ` (${doneCnt})` : ""}</button
    >

    {#if categories.length > 0}
        <span class="text-border text-xs hidden sm:block">|</span>
        {#each categories as cat (cat.id)}
            <button
                onclick={() => {
                    const next = new Set(filterCategory);
                    next.has(cat.id) ? next.delete(cat.id) : next.add(cat.id);
                    filterCategory = next;
                }}
                class="hidden sm:flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs transition-all cursor-pointer {chipCls(
                    filterCategory.has(cat.id),
                )}"
            >
                <span
                    class="w-2 h-2 rounded-full shrink-0"
                    style={dotStyle(cat.color)}
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
                ><UserAvatar name={m.name} avatarUrl={m.avatar_url} userId={m.user_id} size={24} /></button
            >
        {/each}
    {/if}

    {#if someFilterActive}
        <button
            onclick={() => {
                filterTypes = new Set();
                filterMembers = new Set();
                filterCategory = new Set();
            }}
            class="flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            ><X class="w-3 h-3" />{msg.board_clear_filters()}</button
        >
    {/if}
</div>
