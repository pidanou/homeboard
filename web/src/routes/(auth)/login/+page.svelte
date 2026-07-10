<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { base } from '$app/paths';
	import { api, getBaseUrl } from '$lib/api/client';
	import { setToken } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as m from '$lib/paraglide/messages.js';
	import { oidcEnabled, oidcProviderName, allowPasswordLogin } from '$lib/stores/config';

	let email = $state('');
	let password = $state('');
	let remember = $state(true);
	let loading = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			const res = await api.post<{ token: string }>('/api/v1/auth/login', { email, password });
			setToken(res.token, remember);
			goto($page.url.searchParams.get('redirect') ?? `${base}/`);
		} catch { } finally {
			loading = false;
		}
	}

	function continueWithOidc() {
		const redirect = $page.url.searchParams.get('redirect');
		if (redirect) {
			sessionStorage.setItem('oidc_redirect', redirect);
		} else {
			sessionStorage.removeItem('oidc_redirect');
		}
		window.location.href = `${getBaseUrl()}/api/v1/auth/oidc/login`;
	}
</script>

{#if $allowPasswordLogin}
	<form onsubmit={submit} class="flex flex-col gap-4">
		<div class="flex flex-col gap-1.5">
			<Label for="email">{m.auth_email_label()}</Label>
			<Input id="email" type="email" bind:value={email} required />
		</div>
		<div class="flex flex-col gap-1.5">
			<Label for="password">{m.auth_password_label()}</Label>
			<Input id="password" type="password" bind:value={password} required />
		</div>
		<label class="flex items-center gap-2 cursor-pointer">
			<Checkbox bind:checked={remember} />
			<span class="text-sm text-muted-foreground">{m.auth_remember_me()}</span>
		</label>
		<Button type="submit" disabled={loading} class="w-full">
			{loading ? m.auth_signing_in() : m.auth_sign_in()}
		</Button>
		<p class="text-sm text-center text-muted-foreground">
			{m.auth_no_account()} <a href="{base}/register" class="text-primary underline-offset-4 hover:underline">{m.auth_register_link()}</a>
		</p>
	</form>
{/if}

{#if $allowPasswordLogin && $oidcEnabled}
	<div class="flex items-center gap-3 my-4">
		<div class="h-px flex-1 bg-border"></div>
		<span class="text-xs uppercase text-muted-foreground">{m.auth_or_divider()}</span>
		<div class="h-px flex-1 bg-border"></div>
	</div>
{/if}

{#if $oidcEnabled}
	<Button type="button" variant="outline" class="w-full" onclick={continueWithOidc}>
		{m.auth_continue_with({ provider: $oidcProviderName })}
	</Button>
{/if}
