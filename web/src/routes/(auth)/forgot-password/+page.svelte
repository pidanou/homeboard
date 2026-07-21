<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
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

	let email = $state('');
	let loading = $state(false);
	let sent = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			await api.post('/api/v1/auth/forgot-password', { email });
			sent = true;
		} catch {
		} finally {
			loading = false;
		}
	}
</script>

{#if sent}
	<p class="text-sm text-center text-muted-foreground">{m.auth_forgot_password_sent()}</p>
{:else}
	<form onsubmit={submit} class="flex flex-col gap-4">
		<p class="text-sm text-muted-foreground">{m.auth_forgot_password_subtitle()}</p>
		<div class="flex flex-col gap-1.5">
			<Label for="email">{m.auth_email_label()}</Label>
			<Input id="email" type="email" bind:value={email} required />
		</div>
		<Button type="submit" disabled={loading} class="w-full">
			{loading ? m.auth_sending_reset_link() : m.auth_send_reset_link()}
		</Button>
	</form>
{/if}

<p class="text-sm text-center text-muted-foreground mt-4">
	<a href="{base}/login" class="text-primary underline-offset-4 hover:underline">{m.auth_callback_back_to_login()}</a>
</p>
