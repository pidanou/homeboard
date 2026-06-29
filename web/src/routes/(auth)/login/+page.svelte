<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { setToken } from '$lib/auth';
	import { isLocal } from '$lib/env';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';

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
			goto($page.url.searchParams.get('redirect') ?? '/');
		} catch { } finally {
			loading = false;
		}
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-4">
	<div class="flex flex-col gap-1.5">
		<Label for="email">Email</Label>
		<Input id="email" type="email" bind:value={email} required />
	</div>
	<div class="flex flex-col gap-1.5">
		<Label for="password">Password</Label>
		<Input id="password" type="password" bind:value={password} required />
	</div>
	<label class="flex items-center gap-2 cursor-pointer">
		<Checkbox bind:checked={remember} />
		<span class="text-sm text-muted-foreground">Remember me</span>
	</label>
	<Button type="submit" disabled={loading} class="w-full">
		{loading ? 'Signing in…' : 'Sign in'}
	</Button>
	<p class="text-sm text-center text-muted-foreground">
		No account? <a href="/register" class="text-primary underline-offset-4 hover:underline">Register</a>
	</p>
	{#if isLocal}
	<p class="text-sm text-center text-muted-foreground">
		<a href="/setup" class="text-muted-foreground underline-offset-4 hover:underline">Change server</a>
	</p>
	{/if}
</form>
