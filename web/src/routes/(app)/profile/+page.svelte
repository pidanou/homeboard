<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { currentUser, loadCurrentUser } from '$lib/stores/user';
	import UserAvatar from '$lib/components/UserAvatar.svelte';
	import AvatarCrop from '$lib/components/AvatarCrop.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import * as ToggleGroup from '$lib/components/ui/toggle-group';
	import { logout } from '$lib/auth';
	import { subscribePush, unsubscribePush, isPushSubscribed } from '$lib/push';
	import { Camera, ChevronLeft, LogOut, Bell, BellOff } from 'lucide-svelte';
	import * as msg from '$lib/paraglide/messages.js';
	import { getLocale, setLocale, locales } from '$lib/paraglide/runtime';

	let user = $derived($currentUser);

	// Which sub-form is open
	type Panel = null | 'name' | 'password';
	let panel = $state<Panel>(null);

	const LANGUAGE_LABELS: Record<string, string> = { en: 'English', fr: 'Français', es: 'Español', de: 'Deutsch' };

	// Language
	let languageValue = $state(getLocale());
	let languageLoading = $state(false);

	// Time format
	const TIME_FORMAT_LABELS: Record<string, () => string> = {
		auto: msg.profile_time_format_auto,
		'12': msg.profile_time_format_12h,
		'24': msg.profile_time_format_24h,
	};
	let timeFormatValue = $state<'auto' | '12' | '24'>('auto');
	let timeFormatLoading = $state(false);

	// Name edit
	let nameValue = $state('');
	let nameLoading = $state(false);
	let nameError = $state('');

	// Password change
	let currentPw = $state('');
	let newPw = $state('');
	let confirmPw = $state('');
	let pwLoading = $state(false);
	let pwError = $state('');
	let pwSuccess = $state(false);

	// Avatar
	let cropOpen = $state(false);
	let cropComponent = $state<AvatarCrop>(null!);
	let fileInput = $state<HTMLInputElement>(null!);
	let pushSubscribed = $state(false);
	let pushSupported = $state(false);
	let avatarLoading = $state(false);

	onMount(async () => {
		loadCurrentUser();
		pushSupported = 'serviceWorker' in navigator && 'PushManager' in window;
		if (pushSupported) {
			pushSubscribed = await isPushSubscribed();
		}
	});

	$effect(() => {
		if (panel === 'name' && user) nameValue = user.name;
	});

	$effect(() => {
		if (user) timeFormatValue = user.time_format;
	});

	async function saveName() {
		nameError = '';
		if (!nameValue.trim()) { nameError = msg.profile_name_required(); return; }
		nameLoading = true;
		try {
			await api.patch('/api/v1/profile', { name: nameValue.trim() });
			await loadCurrentUser();
			panel = null;
		} catch (e: unknown) {
			nameError = e instanceof Error ? e.message : msg.profile_update_failed();
		} finally {
			nameLoading = false;
		}
	}

	async function savePassword() {
		pwError = '';
		pwSuccess = false;
		if (newPw.length < 8) { pwError = msg.profile_password_min_length(); return; }
		if (newPw !== confirmPw) { pwError = msg.profile_passwords_mismatch(); return; }
		pwLoading = true;
		try {
			await api.patch('/api/v1/profile/password', { current_password: currentPw, new_password: newPw });
			pwSuccess = true;
			currentPw = ''; newPw = ''; confirmPw = '';
			setTimeout(() => { pwSuccess = false; panel = null; }, 1500);
		} catch (e: unknown) {
			pwError = e instanceof Error ? e.message : msg.profile_update_failed();
		} finally {
			pwLoading = false;
		}
	}

	function onFileSelected(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		cropOpen = true;
		// Give the dialog a tick to mount before loading
		setTimeout(() => cropComponent.loadFile(file), 50);
		fileInput.value = '';
	}

	async function onCropConfirm(blob: Blob) {
		avatarLoading = true;
		try {
			const fd = new FormData();
			fd.append('avatar', blob, 'avatar.jpg');
			await api.upload('/api/v1/profile/avatar', fd);
			await loadCurrentUser();
		} finally {
			avatarLoading = false;
		}
	}

	async function saveLanguage(value: string) {
		if (!locales.includes(value as (typeof locales)[number]) || value === getLocale()) return;
		languageValue = value as (typeof locales)[number];
		languageLoading = true;
		try {
			await api.patch('/api/v1/profile/locale', { locale: value });
		} finally {
			languageLoading = false;
		}
		setLocale(value as (typeof locales)[number]);
	}

	async function saveTimeFormat(value: string) {
		if (!['auto', '12', '24'].includes(value) || value === timeFormatValue) return;
		const previous = timeFormatValue;
		timeFormatValue = value as 'auto' | '12' | '24';
		timeFormatLoading = true;
		try {
			await api.patch('/api/v1/profile/time-format', { time_format: value });
			await loadCurrentUser();
		} catch {
			timeFormatValue = previous;
		} finally {
			timeFormatLoading = false;
		}
	}

	async function removeAvatar() {
		avatarLoading = true;
		try {
			await api.delete('/api/v1/profile/avatar');
			await loadCurrentUser();
		} finally {
			avatarLoading = false;
		}
	}
</script>

<div class="px-4 md:px-6 pt-4 md:pt-6 pb-12">
	<div class="max-w-lg mx-auto flex flex-col gap-0 divide-y divide-border">

		<!-- Header -->
		<div class="pb-6">
			<h1 class="text-2xl font-bold">{msg.profile_title()}</h1>
		</div>

		{#if user}
			<!-- Avatar section -->
			<section class="py-6 flex flex-col items-center gap-3">
				<div class="relative">
					<UserAvatar name={user.name} avatarUrl={user.avatar_url} userId={user.id} size={80} />
					{#if avatarLoading}
						<div class="absolute inset-0 rounded-full bg-black/30 flex items-center justify-center">
							<div class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
						</div>
					{/if}
				</div>

				<p class="text-base font-semibold">{user.name}</p>
				<p class="text-sm text-muted-foreground -mt-2">{user.email}</p>

				<div class="flex items-center gap-2">
					<Button variant="outline" size="sm" onclick={() => fileInput.click()} disabled={avatarLoading}>
						<Camera class="w-4 h-4 mr-1.5" />
						{user.avatar_url ? msg.profile_change_photo() : msg.profile_upload_photo()}
					</Button>
					{#if user.avatar_url}
						<Button variant="ghost" size="sm" onclick={removeAvatar} disabled={avatarLoading} class="text-destructive hover:text-destructive">
							{msg.action_remove()}
						</Button>
					{/if}
				</div>
			</section>

			<input bind:this={fileInput} type="file" accept="image/*" class="hidden" onchange={onFileSelected} />
			<AvatarCrop bind:this={cropComponent} bind:open={cropOpen} onconfirm={onCropConfirm} />

			<!-- Account section -->
			<section class="py-6 flex flex-col gap-4">
				<h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{msg.profile_account()}</h2>

				<!-- Display name row -->
				<div class="rounded-xl border border-border overflow-hidden">
					<button
						onclick={() => panel = panel === 'name' ? null : 'name'}
						class="w-full flex items-center justify-between px-4 py-3.5 text-sm hover:bg-muted/50 transition-colors text-left"
					>
						<span class="font-medium">{msg.profile_display_name()}</span>
						<span class="text-muted-foreground text-sm">{user.name}</span>
					</button>

					{#if panel === 'name'}
						<Separator />
						<div class="px-4 py-4 space-y-3 bg-muted/20">
							<div class="space-y-1.5">
								<Label for="display-name">{msg.profile_new_name_label()}</Label>
								<Input
									id="display-name"
									bind:value={nameValue}
									maxlength={64}
									onkeydown={(e) => { if (e.key === 'Enter') saveName(); if (e.key === 'Escape') panel = null; }}
									autofocus
								/>
							</div>
							{#if nameError}<p class="text-sm text-destructive">{nameError}</p>{/if}
							<div class="flex gap-2 justify-end">
								<Button variant="outline" size="sm" onclick={() => panel = null}>{msg.dialog_cancel()}</Button>
								<Button size="sm" onclick={saveName} disabled={nameLoading}>{msg.edit_dialog_save()}</Button>
							</div>
						</div>
					{/if}
				</div>

				<!-- Password row -->
				<div class="rounded-xl border border-border overflow-hidden">
					<button
						onclick={() => panel = panel === 'password' ? null : 'password'}
						class="w-full flex items-center justify-between px-4 py-3.5 text-sm hover:bg-muted/50 transition-colors text-left"
					>
						<span class="font-medium">{msg.profile_change_password()}</span>
						<span class="text-muted-foreground">••••••••</span>
					</button>

					{#if panel === 'password'}
						<Separator />
						<div class="px-4 py-4 space-y-3 bg-muted/20">
							<div class="space-y-1.5">
								<Label for="current-pw">{msg.profile_current_password()}</Label>
								<Input id="current-pw" type="password" bind:value={currentPw} autofocus />
							</div>
							<div class="space-y-1.5">
								<Label for="new-pw">{msg.profile_new_password()}</Label>
								<Input id="new-pw" type="password" bind:value={newPw} />
							</div>
							<div class="space-y-1.5">
								<Label for="confirm-pw">{msg.profile_confirm_password()}</Label>
								<Input id="confirm-pw" type="password" bind:value={confirmPw}
									onkeydown={(e) => { if (e.key === 'Enter') savePassword(); }}
								/>
							</div>
							{#if pwError}<p class="text-sm text-destructive">{pwError}</p>{/if}
							{#if pwSuccess}<p class="text-sm text-green-600">{msg.profile_password_updated()}</p>{/if}
							<div class="flex gap-2 justify-end">
								<Button variant="outline" size="sm" onclick={() => panel = null}>{msg.dialog_cancel()}</Button>
								<Button size="sm" onclick={savePassword} disabled={pwLoading}>{msg.profile_update_button()}</Button>
							</div>
						</div>
					{/if}
				</div>

				<!-- Language row -->
				<div class="rounded-xl border border-border px-4 py-3 flex items-center justify-between gap-3">
					<span class="font-medium text-sm">{msg.profile_language()}</span>
					<ToggleGroup.Root
						type="single"
						bind:value={languageValue}
						onValueChange={(v) => v && saveLanguage(v)}
						variant="outline"
						size="sm"
					>
						{#each locales as loc}
							<ToggleGroup.Item
								value={loc}
								disabled={languageLoading}
								title={LANGUAGE_LABELS[loc] ?? loc}
								class="text-xs font-semibold px-3 data-[state=on]:bg-foreground data-[state=on]:text-background data-[state=on]:hover:bg-foreground/90"
							>
								{loc.toUpperCase()}
							</ToggleGroup.Item>
						{/each}
					</ToggleGroup.Root>
				</div>

				<!-- Time format row -->
				<div class="rounded-xl border border-border px-4 py-3 flex items-center justify-between gap-3">
					<span class="font-medium text-sm">{msg.profile_time_format()}</span>
					<ToggleGroup.Root
						type="single"
						bind:value={timeFormatValue}
						onValueChange={(v) => v && saveTimeFormat(v)}
						variant="outline"
						size="sm"
					>
						{#each ['auto', '12', '24'] as fmt}
							<ToggleGroup.Item
								value={fmt}
								disabled={timeFormatLoading}
								class="text-xs font-semibold px-3 data-[state=on]:bg-foreground data-[state=on]:text-background data-[state=on]:hover:bg-foreground/90"
							>
								{TIME_FORMAT_LABELS[fmt]()}
							</ToggleGroup.Item>
						{/each}
					</ToggleGroup.Root>
				</div>
			</section>
		{:else}
			<div class="flex justify-center py-16">
				<div class="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
			</div>
		{/if}

		{#if pushSupported}
			<section class="py-6 flex flex-col gap-4">
				<h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{msg.profile_notifications()}</h2>
				<Button
					variant={pushSubscribed ? 'outline' : 'default'}
					onclick={async () => {
						try {
							if (pushSubscribed) {
								await unsubscribePush();
								pushSubscribed = false;
							} else {
								await subscribePush();
								pushSubscribed = await isPushSubscribed();
							}
						} catch (e) {
							console.error('push error:', e);
						}
					}}
				>
					{#if pushSubscribed}
						<BellOff class="w-4 h-4 mr-2" /> {msg.profile_disable_notifications()}
					{:else}
						<Bell class="w-4 h-4 mr-2" /> {msg.profile_enable_notifications()}
					{/if}
				</Button>
			</section>
		{/if}

		<section class="py-6">
			<Button variant="destructive" class="w-full" onclick={logout}>
				<LogOut class="w-4 h-4 mr-2" />
				{msg.sidebar_sign_out()}
			</Button>
		</section>

	</div>
</div>
