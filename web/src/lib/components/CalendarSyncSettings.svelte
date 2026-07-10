<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api/client";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { X, RefreshCw } from "lucide-svelte";
    import * as msg from "$lib/paraglide/messages.js";
    import type { CalendarExportToken, CalendarSubscription } from "$lib/types";

    let { familyID, isAdmin }: { familyID: string | undefined; isAdmin: boolean } = $props();

    let exportToken = $state<CalendarExportToken | null>(null);
    let subscriptions = $state<CalendarSubscription[]>([]);
    let copied = $state(false);
    let importing = $state(false);
    let importResult = $state<{ imported: number; skipped: number } | null>(null);
    let importFileInput = $state<HTMLInputElement>(null!);
    let addingSubscription = $state(false);
    let newSubName = $state("");
    let newSubUrl = $state("");
    let syncingId = $state<string | null>(null);

    onMount(async () => {
        if (!isAdmin || !familyID) return;
        const [tokenResult, subsResult] = await Promise.allSettled([
            api.get<CalendarExportToken | null>(
                `/api/v1/households/${familyID}/calendar/export-token`,
            ),
            api.get<CalendarSubscription[]>(
                `/api/v1/households/${familyID}/calendar/subscriptions`,
            ),
        ]);
        if (tokenResult.status === "fulfilled") exportToken = tokenResult.value;
        if (subsResult.status === "fulfilled") subscriptions = subsResult.value ?? [];
    });

    function feedUrl(token: string) {
        return `${location.origin}/api/v1/calendar/export/${token}.ics`;
    }

    async function generateExportToken() {
        try {
            exportToken = await api.post<CalendarExportToken>(
                `/api/v1/households/${familyID}/calendar/export-token`,
                {},
            );
        } catch {}
    }

    async function revokeExportToken() {
        try {
            await api.delete(`/api/v1/households/${familyID}/calendar/export-token`);
            exportToken = null;
        } catch {}
    }

    function copyExportLink() {
        if (!exportToken) return;
        navigator.clipboard.writeText(feedUrl(exportToken.token));
        copied = true;
        setTimeout(() => (copied = false), 2000);
    }

    async function onImportFileSelected(e: Event) {
        const file = (e.target as HTMLInputElement).files?.[0];
        importFileInput.value = "";
        if (!file) return;
        importing = true;
        importResult = null;
        try {
            const fd = new FormData();
            fd.append("file", file);
            importResult = await api.upload<{ imported: number; skipped: number }>(
                `/api/v1/households/${familyID}/calendar/import`,
                fd,
            );
        } finally {
            importing = false;
        }
    }

    async function addSubscription() {
        if (!newSubName.trim() || !newSubUrl.trim()) return;
        try {
            const sub = await api.post<CalendarSubscription>(
                `/api/v1/households/${familyID}/calendar/subscriptions`,
                { name: newSubName.trim(), url: newSubUrl.trim() },
            );
            subscriptions = [...subscriptions, sub];
            newSubName = "";
            newSubUrl = "";
            addingSubscription = false;
        } catch {}
    }

    async function deleteSubscription(id: string) {
        try {
            await api.delete(`/api/v1/households/${familyID}/calendar/subscriptions/${id}`);
            subscriptions = subscriptions.filter((s) => s.id !== id);
        } catch {}
    }

    async function syncNow(id: string) {
        syncingId = id;
        try {
            await api.post(`/api/v1/households/${familyID}/calendar/subscriptions/${id}/sync`, {});
        } catch {
        } finally {
            const subs = await api
                .get<CalendarSubscription[]>(`/api/v1/households/${familyID}/calendar/subscriptions`)
                .catch(() => null);
            if (subs) subscriptions = subs;
            syncingId = null;
        }
    }
</script>

{#if isAdmin}
    <!-- Export -->
    <section class="py-4 flex flex-col gap-4">
        <div class="flex items-center justify-between gap-3">
            <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {msg.settings_calendar_export()}
            </h2>
            <Button size="sm" variant="outline" onclick={generateExportToken}>
                {exportToken ? msg.settings_regenerate() : msg.settings_generate_link()}
            </Button>
        </div>

        {#if exportToken}
            <div class="rounded-xl border border-border bg-card overflow-hidden">
                <div class="flex items-center gap-2 px-4 py-3 text-xs font-mono text-muted-foreground border-b border-border">
                    <span class="flex-1 truncate">{feedUrl(exportToken.token)}</span>
                </div>
                <div class="flex gap-2 px-4 py-3">
                    <Button variant="outline" size="sm" class="flex-1" onclick={copyExportLink}>
                        {copied ? msg.settings_copied() : msg.settings_copy_link()}
                    </Button>
                    <Button variant="destructive" size="sm" onclick={revokeExportToken}>
                        {msg.settings_revoke()}
                    </Button>
                </div>
            </div>
        {:else}
            <p class="text-sm text-muted-foreground">{msg.settings_calendar_no_export()}</p>
        {/if}
    </section>

    <!-- Import -->
    <section class="py-4 flex flex-col gap-4">
        <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {msg.settings_calendar_import()}
        </h2>
        <div class="flex items-center gap-2">
            <Button
                size="sm"
                variant="outline"
                disabled={importing}
                onclick={() => importFileInput.click()}
            >
                {importing ? msg.settings_calendar_importing() : msg.settings_calendar_import_file()}
            </Button>
            {#if importResult}
                <span class="text-xs text-muted-foreground">
                    {msg.settings_calendar_import_result({
                        imported: importResult.imported,
                        skipped: importResult.skipped,
                    })}
                </span>
            {/if}
        </div>
        <input
            bind:this={importFileInput}
            type="file"
            accept=".ics,text/calendar"
            class="hidden"
            onchange={onImportFileSelected}
        />
    </section>

    <!-- Subscriptions -->
    <section class="py-4 flex flex-col gap-4">
        <div class="flex items-center justify-between gap-3">
            <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {msg.settings_calendar_subscriptions()}
            </h2>
            <Button size="sm" variant="outline" onclick={() => (addingSubscription = !addingSubscription)}>
                {msg.action_add()}
            </Button>
        </div>

        {#if addingSubscription}
            <div class="flex flex-col gap-2">
                <Input
                    bind:value={newSubName}
                    placeholder={msg.settings_calendar_subscription_name_placeholder()}
                    class="h-8 text-sm"
                />
                <Input
                    bind:value={newSubUrl}
                    placeholder={msg.settings_calendar_subscription_url_placeholder()}
                    class="h-8 text-sm"
                    onkeydown={(e) => {
                        if (e.key === "Enter") {
                            e.preventDefault();
                            addSubscription();
                        }
                    }}
                />
                <div class="flex gap-2">
                    <Button
                        size="sm"
                        onclick={addSubscription}
                        disabled={!newSubName.trim() || !newSubUrl.trim()}
                        class="h-8">{msg.action_add()}</Button
                    >
                    <Button size="sm" variant="ghost" class="h-8" onclick={() => (addingSubscription = false)}>
                        {msg.dialog_cancel()}
                    </Button>
                </div>
            </div>
        {/if}

        {#if subscriptions.length === 0}
            <p class="text-sm text-muted-foreground">{msg.settings_calendar_no_subscriptions()}</p>
        {:else}
            <div class="rounded-xl border border-border bg-card overflow-hidden divide-y divide-border">
                {#each subscriptions as sub (sub.id)}
                    <div class="flex items-center gap-3 px-4 py-3">
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium truncate">{sub.name}</p>
                            <p class="text-xs text-muted-foreground truncate">
                                {#if sub.last_sync_error}
                                    <span class="text-destructive">{sub.last_sync_error}</span>
                                {:else if sub.last_synced_at}
                                    {msg.settings_calendar_last_synced({
                                        time: new Date(sub.last_synced_at).toLocaleString(),
                                    })}
                                {:else}
                                    {msg.settings_calendar_never_synced()}
                                {/if}
                            </p>
                        </div>
                        <button
                            onclick={() => syncNow(sub.id)}
                            disabled={syncingId === sub.id}
                            class="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                            aria-label={msg.settings_calendar_sync_now()}
                        >
                            <RefreshCw class="w-3.5 h-3.5 {syncingId === sub.id ? 'animate-spin' : ''}" />
                        </button>
                        <button
                            onclick={() => deleteSubscription(sub.id)}
                            class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                            aria-label={msg.action_remove()}
                        >
                            <X class="w-4 h-4" />
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </section>
{/if}
