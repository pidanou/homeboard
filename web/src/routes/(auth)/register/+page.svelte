<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { setToken } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';

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
			goto('/');
		} catch { } finally {
			loading = false;
		}
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-4">
	<div class="flex flex-col gap-1.5">
		<Label for="name">Name</Label>
		<Input id="name" type="text" bind:value={name} required />
	</div>
	<div class="flex flex-col gap-1.5">
		<Label for="email">Email</Label>
		<Input id="email" type="email" bind:value={email} required />
	</div>
	<div class="flex flex-col gap-1.5">
		<Label for="password">Password</Label>
		<Input id="password" type="password" bind:value={password} required minlength={8} />
	</div>
<Button type="submit" disabled={loading} class="w-full">
		{loading ? 'Creating account…' : 'Create account'}
	</Button>
	<p class="text-sm text-center text-muted-foreground">
		Already have an account? <a href="/login" class="text-primary underline-offset-4 hover:underline">Sign in</a>
	</p>
</form>
