<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';

	let url = $state('');
	let error = $state('');

	function save() {
		error = '';
		let trimmed = url.trim().replace(/\/$/, '');
		try {
			new URL(trimmed);
		} catch {
			error = 'Enter a valid URL, e.g. https://household.example.com';
			return;
		}
		localStorage.setItem('api_url', trimmed);
		goto('/');
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-background px-4 py-12">
	<div class="w-full max-w-sm">
		<div class="text-center mb-8">
			<span class="text-4xl">🏠</span>
			<h1 class="text-2xl font-bold mt-2 text-foreground">Homeboard</h1>
			<p class="text-sm text-muted-foreground mt-1">Enter your Homeboard server address to get started.</p>
		</div>
		<div class="bg-card border border-border rounded-xl p-6 shadow-sm flex flex-col gap-4">
			<div class="flex flex-col gap-1.5">
				<Label for="api-url">Server URL</Label>
				<Input
					id="api-url"
					bind:value={url}
					placeholder="https://household.example.com"
					type="url"
					autocapitalize="none"
					autocorrect="off"
				/>
				{#if error}
					<p class="text-destructive text-xs">{error}</p>
				{/if}
			</div>
			<Button onclick={save} disabled={!url.trim()}>Connect</Button>
		</div>
	</div>
</div>
