<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { base } from '$app/paths';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as m from '$lib/paraglide/messages.js';
	import { allowPasswordLogin } from '$lib/stores/config';

	onMount(() => {
		const unsubscribe = allowPasswordLogin.subscribe((allowed) => {
			if (!allowed) goto(`${base}/login`);
		});
		return unsubscribe;
	});

	const token = $derived($page.params.token);

	let password = $state('');
	let loading = $state(false);
	let done = $state(false);
	let error = $state('');

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			await api.post('/api/v1/auth/reset-password', { token, password });
			done = true;
		} catch {
			error = m.auth_reset_password_invalid();
		} finally {
			loading = false;
		}
	}
</script>

{#if done}
	<p class="text-sm text-center text-muted-foreground">{m.auth_reset_password_success()}</p>
	<p class="text-sm text-center mt-4">
		<a href="{base}/login" class="text-primary underline-offset-4 hover:underline">{m.auth_callback_back_to_login()}</a>
	</p>
{:else}
	<form onsubmit={submit} class="flex flex-col gap-4">
		<h2 class="text-lg font-semibold text-center">{m.auth_reset_password_title()}</h2>
		<div class="flex flex-col gap-1.5">
			<Label for="password">{m.auth_new_password_label()}</Label>
			<Input id="password" type="password" bind:value={password} required minlength={8} />
		</div>
		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}
		<Button type="submit" disabled={loading} class="w-full">
			{loading ? m.auth_resetting_password() : m.auth_reset_password_button()}
		</Button>
	</form>
{/if}
