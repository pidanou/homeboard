<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';

	let name = $state('');
	let error = $state('');
	let loading = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		error = '';
		try {
			const family = await api.post<{ id: string }>('/api/v1/families', { name });
			goto(`/families/${family.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create family';
		} finally {
			loading = false;
		}
	}
</script>

<div class="max-w-sm mx-auto">
	<h2 class="text-xl font-semibold mb-6">Create a family</h2>
	<form onsubmit={submit} class="flex flex-col gap-4">
		<div class="flex flex-col gap-1">
			<label for="name" class="text-sm font-medium">Family name</label>
			<input
				id="name"
				type="text"
				bind:value={name}
				required
				placeholder="The Smiths"
				class="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
			/>
		</div>
		{#if error}
			<p class="text-red-500 text-sm">{error}</p>
		{/if}
		<button
			type="submit"
			disabled={loading}
			class="bg-primary-500 text-white rounded-lg px-4 py-2 font-medium hover:bg-primary-600 disabled:opacity-50 transition-colors"
		>
			{loading ? 'Creating…' : 'Create'}
		</button>
	</form>
</div>
