<script lang="ts">
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
	import { api } from '$lib/api/client';
	import { setToken } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as m from '$lib/paraglide/messages.js';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let loading = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			await api.post('/api/v1/auth/register', { name, email, password });
			const res = await api.post<{ token: string }>('/api/v1/auth/login', { email, password });
			setToken(res.token);
			goto(`${base}/`);
		} catch { } finally {
			loading = false;
		}
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-4">
	<div class="flex flex-col gap-1.5">
		<Label for="name">{m.auth_name_label()}</Label>
		<Input id="name" type="text" bind:value={name} required />
	</div>
	<div class="flex flex-col gap-1.5">
		<Label for="email">{m.auth_email_label()}</Label>
		<Input id="email" type="email" bind:value={email} required />
	</div>
	<div class="flex flex-col gap-1.5">
		<Label for="password">{m.auth_password_label()}</Label>
		<Input id="password" type="password" bind:value={password} required minlength={8} />
	</div>
<Button type="submit" disabled={loading} class="w-full">
		{loading ? m.auth_creating_account() : m.auth_create_account()}
	</Button>
	<p class="text-sm text-center text-muted-foreground">
		{m.auth_have_account()} <a href="{base}/login" class="text-primary underline-offset-4 hover:underline">{m.auth_sign_in_link()}</a>
	</p>
</form>
