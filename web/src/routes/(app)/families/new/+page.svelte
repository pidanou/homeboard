<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';

	let name = $state('');
	let loading = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			const family = await api.post<{ id: string }>('/api/v1/families', { name });
			goto(`/families/${family.id}`);
		} catch { } finally {
			loading = false;
		}
	}
</script>

<div class="px-4 md:px-6 pt-6 pb-8">
	<div class="max-w-sm mx-auto">
		<h2 class="text-xl font-semibold mb-6">Create a household</h2>
		<form onsubmit={submit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-1.5">
				<Label for="name">Household name</Label>
				<Input id="name" bind:value={name} required placeholder="The Smiths" />
			</div>
			<Button type="submit" disabled={loading}>
				{loading ? 'Creating…' : 'Create'}
			</Button>
		</form>
	</div>
</div>
