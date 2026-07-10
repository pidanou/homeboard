<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { base } from '$app/paths';
	import { api } from '$lib/api/client';
	import { setToken } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import { oidcProviderName } from '$lib/stores/config';
	import * as m from '$lib/paraglide/messages.js';

	let errorReason = $state<string | null>(null);

	onMount(async () => {
		const code = $page.url.searchParams.get('code');
		const error = $page.url.searchParams.get('error');

		if (error) {
			errorReason = error;
			return;
		}
		if (!code) {
			errorReason = 'oidc_failed';
			return;
		}

		try {
			const res = await api.post<{ token: string }>('/api/v1/auth/oidc/exchange', { code });
			setToken(res.token, true);
			const redirect = sessionStorage.getItem('oidc_redirect');
			sessionStorage.removeItem('oidc_redirect');
			goto(redirect ?? `${base}/`);
		} catch {
			errorReason = 'oidc_failed';
		}
	});

	function errorMessage(reason: string): string {
		if (reason === 'email_not_verified') {
			return m.auth_callback_error_email_not_verified({ provider: $oidcProviderName });
		}
		if (reason === 'registration_closed') {
			return m.auth_callback_error_registration_closed();
		}
		return m.auth_callback_error_generic();
	}
</script>

{#if errorReason}
	<div class="flex flex-col gap-4 items-center text-center">
		<p class="text-sm text-destructive">{errorMessage(errorReason)}</p>
		<Button href="{base}/login" class="w-full">{m.auth_callback_back_to_login()}</Button>
	</div>
{:else}
	<p class="text-sm text-center text-muted-foreground">{m.auth_callback_signing_in()}</p>
{/if}
